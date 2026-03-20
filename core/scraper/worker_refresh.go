// core/scraper/worker_refresh.go
package scraper

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
)

// runRefreshWorker is the per-agent goroutine for refresh mode.
// It launches the browser, then pulls accounts from the channel one by one,
// clearing the session between each to ensure a clean login.
func (s *Scraper) runRefreshWorker(
	ctx context.Context,
	agent *ScraperAgent,
	accounts <-chan refreshAccountEntry,
	results chan<- refreshResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if !s.launchBrowser(ctx, agent) {
		return
	}

	s.emitEvent(EventAgentStarted, agent.Name, "refresh worker ready", nil)
	defer s.shutdownWorker(agent)

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-accounts:
			if !ok {
				agent.SetStatusText(i18n.T("agent.allAccountsRefreshed"))
				return
			}

			agent.SetStatus(AgentStatusBusy)
			agent.SetStatusText(i18n.T("agent.clearingSession"))
			if err := agent.ClearSession(); err != nil {
				logger.Warn().Err(err).Str("agent", agent.Name).Msg("failed to clear session between accounts")
			}

			result := s.refreshAccount(ctx, agent, entry)
			if ctx.Err() != nil {
				return
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// refreshAccount logs into a single account and fetches its live credit balance.
// Session clearing between accounts is handled by the worker loop.
func (s *Scraper) refreshAccount(ctx context.Context, agent *ScraperAgent, entry refreshAccountEntry) refreshResult {
	account := entry.Account
	if ctx.Err() != nil {
		return refreshResult{Email: account.Email, Error: ctx.Err(), Agent: agent.Name}
	}

	logger.Info().Str("agent", agent.Name).Str("email", account.Email).Msg("refreshing account credits")

	fail := func(err error) refreshResult {
		return refreshResult{Email: account.Email, Error: err, Agent: agent.Name}
	}

	login := func() error {
		agent.SetStatusText(i18n.T("agent.loggingIn"))
		return s.loginWithRetry(ctx, agent, account)
	}

	fetchCredits := func() error {
		return s.bootstrapCredits(ctx, agent)
	}

	if err := login(); err != nil {
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}

		if isConnectionError(err) {
			s.relaunchOnConnectionError(ctx, agent, err)
		}

		// Re-queue for later if under the retry limit.
		if !errors.Is(err, ErrAccountNotActive) && entry.Retries < maxRefreshRetries {
			logger.Warn().Err(err).Str("agent", agent.Name).Str("email", account.Email).
				Int("retries", entry.Retries).Msg("refresh login failed (will retry later)")
			agent.SetStatusText(i18n.T("agent.loginFailedRetry"))
			return refreshResult{
				Email:   account.Email,
				Agent:   agent.Name,
				Account: account,
				Retry:   true,
				Retries: entry.Retries + 1,
			}
		}

		logger.Warn().Err(err).Str("agent", agent.Name).Str("email", account.Email).Msg("refresh login failed")
		agent.SetStatusText(i18n.T("agent.loginFailed"))
		s.emitEvent(EventAgentError, agent.Name, fmt.Sprintf("login failed: %s", account.Email), map[string]any{
			"email": account.Email,
			"error": err.Error(),
		})
		return fail(err)
	}

	if err := fetchCredits(); err != nil {
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}
		logger.Warn().Err(err).Str("agent", agent.Name).Str("email", account.Email).Msg("credit refresh failed")
		agent.SetStatusText(i18n.T("agent.creditFetchFailed"))
		s.emitEvent(EventAgentError, agent.Name, fmt.Sprintf("credit refresh failed: %s", account.Email), map[string]any{
			"email": account.Email,
			"error": err.Error(),
		})
		return fail(err)
	}

	if acc := agent.Account(); acc != nil {
		s.refreshedCredits.Add(int32(acc.CreditsAmount))
	}

	s.emitEvent(EventAccountRefreshed, agent.Name, fmt.Sprintf("credits refreshed: %s", account.Email), map[string]any{
		"email": account.Email,
	})

	logger.Info().Str("agent", agent.Name).Str("email", account.Email).Msg("account credits refreshed")

	return refreshResult{Email: account.Email, Agent: agent.Name}
}
