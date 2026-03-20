// core/scraper/worker_scrape_query.go
package scraper

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
)

// recordQueryResult updates the counters and emits the appropriate event.
// Queries never "fail" - they either return hosts (completed) or don't (empty).
// Errors are agent-level issues; they still count as completed for progress
// purposes but emit EventQueryFailed so the UI can see the agent had trouble.
func (s *Scraper) recordQueryResult(result queryResult, completed, empty, hostCount *atomic.Int32) {
	if result.Error != nil {
		// Agent-level error - count as completed (query itself is fine).
		completed.Add(1)
		s.emitEvent(EventQueryFailed, result.Agent, fmt.Sprintf("query failed: %s", result.Query), map[string]any{
			"query": result.Query,
			"error": result.Error.Error(),
		})
	} else if len(result.Hosts) == 0 {
		// Query succeeded but found no hosts.
		empty.Add(1)
		s.emitEvent(EventQueryCompleted, result.Agent, fmt.Sprintf("query completed: 0 hosts"), map[string]any{
			"query":      result.Query,
			"host_count": 0,
		})
	} else {
		completed.Add(1)
		hostCount.Add(int32(len(result.Hosts)))
		s.emitEvent(EventQueryCompleted, result.Agent, fmt.Sprintf("query completed: %d hosts", len(result.Hosts)), map[string]any{
			"query":      result.Query,
			"host_count": len(result.Hosts),
		})
	}
}

// processQuery executes a single query, handling credit exhaustion by
// swapping the agent's account if needed.
func (s *Scraper) processQuery(ctx context.Context, agent *ScraperAgent, query string) (queryResult, *ScraperAgent) {
	if ctx.Err() != nil {
		return queryResult{Query: query, Error: ctx.Err(), Agent: agent.Name}, nil
	}

	fail := func(err error) queryResult {
		return queryResult{Query: query, Error: err, Agent: agent.Name}
	}

	if agent.Account() == nil {
		logger.Error().Str("agent", agent.Name).Msg("agent has no account assigned, skipping query")
		agent.SetStatus(AgentStatusError)
		return fail(ErrAgentNotLoggedIn), nil
	}

	prepareAndEnsureCredits := func() error {
		agent.SetStatus(AgentStatusBusy)
		agent.SetStatusText(i18n.T("agent.scraping"))
		s.emitEvent(EventQueryStarted, agent.Name, fmt.Sprintf("starting query: %s", query), map[string]any{
			"query": query,
		})
		logger.Info().Str("agent", agent.Name).Str("query", query).Msg("processing query")
		return s.ensureCredits(ctx, agent)
	}

	if err := prepareAndEnsureCredits(); err != nil {
		if errors.Is(err, ErrRanOutOfCredits) {
			return s.handleCreditExhaustion(ctx, agent, query), nil
		}
		agent.SetStatus(AgentStatusError)
		return fail(fmt.Errorf("ensure credits: %w", err)), nil
	}

	hosts, err := agent.Scrape(ctx, query)
	if err != nil {
		if errors.Is(err, ErrRanOutOfCredits) {
			return s.handleCreditExhaustion(ctx, agent, query), nil
		}
		logger.Warn().Err(err).Str("agent", agent.Name).Str("query", query).Msg("query failed")
		agent.SetStatus(AgentStatusError)
		return fail(err), nil
	}

	if len(hosts) > 0 {
		agent.DeductEstimatedCredits()
		s.usedCredits.Add(CreditsAmountPerQuery)
		s.markCreditsUsed(agent)
	}
	storeScrapedHosts(hosts)

	agent.SetStatus(AgentStatusIdle)
	agent.SetStatusText(i18n.T("agent.queryDone"))
	logger.Info().Str("agent", agent.Name).Str("query", query).Int("hosts", len(hosts)).Int("estimated_credits", agent.EstimatedCredits()).Msg("query processed successfully")

	return queryResult{Query: query, Hosts: hosts, Agent: agent.Name}, nil
}

// handleCreditExhaustion replaces the agent's account and retries the query.
// Returns an error wrapping ErrNoUsableAccounts if no pool account is available.
func (s *Scraper) handleCreditExhaustion(ctx context.Context, agent *ScraperAgent, query string) queryResult {
	logger.Warn().Str("agent", agent.Name).Msg("account ran out of credits, replacing account")
	agent.SetStatusText(i18n.T("agent.outOfCreditsReplacing"))

	if err := s.replaceAccount(ctx, agent); err != nil {
		return queryResult{Query: query, Error: fmt.Errorf("credits exhausted, account replacement failed: %w", err), Agent: agent.Name}
	}

	if ctx.Err() != nil {
		return queryResult{Query: query, Error: ctx.Err(), Agent: agent.Name}
	}

	agent.SetStatus(AgentStatusBusy)
	agent.SetStatusText(i18n.T("agent.retryingQuery"))
	hosts, err := agent.Scrape(ctx, query)
	if err != nil {
		agent.SetStatus(AgentStatusError)
		return queryResult{Query: query, Error: err, Agent: agent.Name}
	}

	if len(hosts) > 0 {
		agent.DeductEstimatedCredits()
		s.usedCredits.Add(CreditsAmountPerQuery)
		s.markCreditsUsed(agent)
	}
	storeScrapedHosts(hosts)

	agent.SetStatus(AgentStatusIdle)
	logger.Info().Str("agent", agent.Name).Str("query", query).Int("hosts", len(hosts)).Int("estimated_credits", agent.EstimatedCredits()).Msg("query processed after account replacement")

	return queryResult{Query: query, Hosts: hosts, Agent: agent.Name}
}

// replaceAccount releases the agent's current account, assigns a new one
// from the pool, and logs in. The same browser/agent is reused.
// Returns ErrNoUsableAccounts if no pool account is available.
func (s *Scraper) replaceAccount(ctx context.Context, agent *ScraperAgent) error {
	s.pool.release(agent.Name)
	_ = agent.ClearSession()
	agent.SetAccountHint("")

	return s.bootstrapAccount(ctx, agent)
}

// ensureCredits checks that the agent has enough credits, fetching live balance if needed.
func (s *Scraper) ensureCredits(ctx context.Context, agent *ScraperAgent) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if agent.EstimatedCredits() >= CreditsAmountPerQuery {
		return nil
	}

	agent.SetStatusText(i18n.T("agent.checkingCredits"))
	logger.Info().Str("agent", agent.Name).Int("estimated", agent.EstimatedCredits()).Msg("estimated credits low, fetching real balance")
	credits, err := agent.UpdateCredits(ctx)
	if err != nil {
		logger.Warn().Err(err).Str("agent", agent.Name).Msg("failed to fetch real credits, treating as zero")
	} else if credits.Current >= CreditsAmountPerQuery {
		return nil
	}

	actual := uint(0)
	if agent.Account() != nil {
		actual = agent.Account().CreditsAmount
	}

	logger.Warn().Str("agent", agent.Name).Uint("credits", actual).Msg("insufficient credits, agent needs replacement")
	s.emitEvent(EventCreditsLow, agent.Name, "credits too low for search", map[string]any{
		"credits": actual,
	})

	return ErrRanOutOfCredits
}

// markCreditsUsed stamps CreditsLastUsedAt on the agent's account, syncs the
// estimated credit balance back to the model, and persists it so the refresh
// run can skip accounts whose credits haven't been touched.
func (s *Scraper) markCreditsUsed(agent *ScraperAgent) {
	account := agent.Account()
	if account == nil {
		return
	}
	now := time.Now()
	account.CreditsLastUsedAt = &now
	account.CreditsAmount = uint(agent.EstimatedCredits())
	s.pool.persistCredits(account)
}
