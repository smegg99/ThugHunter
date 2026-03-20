// core/scraper/run_helpers.go
package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"smegg.me/thughunter/common/config"
	"smegg.me/thughunter/common/logger"
)

// createAgents creates up to count agents for the current run.
func (s *Scraper) createAgents(ctx context.Context, count int) ([]*ScraperAgent, error) {
	logger.Info().Int("count", count).Msg("creating run agents")

	agents := make([]*ScraperAgent, 0, count)
	for range count {
		select {
		case <-ctx.Done():
			s.cleanupRunAgents()
			return nil, ctx.Err()
		default:
		}

		agent, err := s.CreateAgent()
		if err != nil {
			logger.Error().Err(err).Msg("failed to create run agent")
			break
		}

		agents = append(agents, agent)
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("failed to create any run agents")
	}

	logger.Info().Int("count", len(agents)).Msg("run agents created")

	s.emitEvent(EventAgentStarted, "", fmt.Sprintf("%d agents created", len(agents)), map[string]any{
		"agent_count": len(agents),
	})

	return agents, nil
}

// forceCloseGrace is the time to wait after context cancellation before
// force-killing all browsers. This gives workers a brief window to exit
// their current rod calls cleanly.
const forceCloseGrace = 1 * time.Second

// forceCloseOnCancel spawns a goroutine that watches ctx and, once cancelled,
// waits forceCloseGrace then force-kills every agent's browser. This unblocks
// workers stuck in long-running rod operations. Returns a stop func that must
// be called when the run finishes to prevent a leak.
func (s *Scraper) forceCloseOnCancel(ctx context.Context) (stop func()) {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			time.Sleep(forceCloseGrace)
			s.mu.Lock()
			agents := make([]*ScraperAgent, 0, len(s.agents))
			for _, a := range s.agents {
				agents = append(agents, a)
			}
			s.mu.Unlock()
			for _, a := range agents {
				logger.Warn().Str("agent", a.Name).Msg("force-killing browser after cancel grace period")
				a.ForceClose()
			}
		case <-done:
		}
	}()
	return func() { close(done) }
}

// cleanupRunAgents removes all agents from the map and force-closes them
// in parallel. Uses ForceClose for instant browser teardown during run
// cancellation rather than waiting for graceful shutdown.
func (s *Scraper) cleanupRunAgents() {
	s.mu.Lock()
	agents := make(map[string]*ScraperAgent, len(s.agents))
	for k, v := range s.agents {
		agents[k] = v
	}
	s.agents = make(map[string]*ScraperAgent)
	s.mu.Unlock()

	var wg sync.WaitGroup
	for name, agent := range agents {
		wg.Add(1)
		go func(n string, a *ScraperAgent) {
			defer wg.Done()
			a.ForceClose()
		}(name, agent)
	}
	wg.Wait()
}

// buildQueries generates all built-in queries (base + per-country) plus any
// user-defined custom queries from config.
func buildQueries() []string {
	var queries []string

	// Base VNC queries.
	queries = append(queries, string(BaseVNCQueryString))
	for _, country := range AllCountries {
		q, err := ResolveQueryForCountry(BaseVNCByCountryQueryString, country)
		if err != nil {
			logger.Warn().Err(err).Str("country", string(country)).Msg("failed to resolve VNC country query template")
			continue
		}
		queries = append(queries, q)
	}

	// Native VNC queries (non-VM, non-container, non-cloud).
	queries = append(queries, string(BaseVNCNativeQueryString))
	for _, country := range AllCountries {
		q, err := ResolveQueryForCountry(BaseVNCNativeByCountryQueryString, country)
		if err != nil {
			logger.Warn().Err(err).Str("country", string(country)).Msg("failed to resolve VNC native country query template")
			continue
		}
		queries = append(queries, q)
	}

	// Camera queries.
	queries = append(queries, string(BaseCameraNoAuthQueryString))

	builtinCount := 3 + 2*len(AllCountries)

	cfg := config.Get()
	for _, q := range cfg.Scraper.QueryStrings {
		if s, ok := q.(string); ok && s != "" {
			queries = append(queries, expandCustomQuery(s)...)
		}
	}

	logger.Info().
		Int("builtin", builtinCount).
		Int("custom", len(cfg.Scraper.QueryStrings)).
		Int("total", len(queries)).
		Msg("query list built")

	return queries
}

// expandCustomQuery expands a single custom query string. It supports:
//   - {{ALL_COUNTRIES}} - replaced by one copy per country
//   - {{CONTINENT:Name}} - replaced by one copy per country in the named continent
//
// If neither placeholder is present, the query is returned as-is.
func expandCustomQuery(q string) []string {
	if strings.Contains(q, "{{ALL_COUNTRIES}}") {
		out := make([]string, 0, len(AllCountries))
		for _, c := range AllCountries {
			out = append(out, strings.ReplaceAll(q, "{{ALL_COUNTRIES}}", c.String()))
		}
		return out
	}

	const prefix = "{{CONTINENT:"
	if idx := strings.Index(q, prefix); idx >= 0 {
		end := strings.Index(q[idx:], "}}")
		if end > len(prefix) {
			tag := q[idx : idx+end+2]
			name := q[idx+len(prefix) : idx+end]
			countries, ok := CountriesByContinent[Continent(name)]
			if ok && len(countries) > 0 {
				out := make([]string, 0, len(countries))
				for _, c := range countries {
					out = append(out, strings.ReplaceAll(q, tag, c.String()))
				}
				return out
			}
			logger.Warn().Str("continent", name).Msg("unknown continent in custom query template")
		}
	}

	return []string{q}
}

// Progress returns a snapshot of the active run.
func (s *Scraper) Progress() *RunProgress {
	s.mu.Lock()
	agentInfos := make([]AgentInfo, 0, len(s.agents))
	activeAgents := 0
	for _, agent := range s.agents {
		agentInfos = append(agentInfos, agent.Info())
		if agent.status != AgentStatusOffline && agent.status != AgentStatusError {
			activeAgents++
		}
	}
	totalAgents := len(s.agents)
	running := s.running
	mode := s.mode
	startedAt := s.startedAt
	s.mu.Unlock()

	total := int(s.totalAccounts.Load())
	refreshed := int(s.refreshedAccounts.Load())
	failed := int(s.failedAccounts.Load())
	remaining := total - refreshed - failed
	if remaining < 0 {
		remaining = 0
	}

	totalCredits := int(s.refreshedCredits.Load())
	possibleQueries := 0
	if CreditsAmountPerQuery > 0 && totalCredits > 0 {
		possibleQueries = totalCredits / CreditsAmountPerQuery
	}

	return &RunProgress{
		Running:             running,
		Mode:                mode,
		TotalQueries:        int(s.totalQueries.Load()),
		CompletedQueries:    int(s.completedQueries.Load()),
		EmptyQueries:        int(s.emptyQueries.Load()),
		TotalHosts:          int(s.totalHosts.Load()),
		TotalAccounts:       total,
		RefreshedAccounts:   refreshed,
		FailedAccounts:      failed,
		RemainingAccounts:   remaining,
		CreatedAccounts:     int(s.createdAccounts.Load()),
		FailedRegistrations: int(s.failedRegistrations.Load()),
		TargetAccounts:      int(s.targetAccounts.Load()),
		DurationSecs:        int(s.runDurationSecs.Load()),
		ActiveAgents:        activeAgents,
		TotalAgents:         totalAgents,
		TotalCredits:        totalCredits,
		PossibleQueries:     possibleQueries,
		UsedCredits:         int(s.usedCredits.Load()),
		AccountsExhausted:   s.accountsExhausted.Load(),
		StartedAt:           startedAt,
		Agents:              agentInfos,
	}
}
