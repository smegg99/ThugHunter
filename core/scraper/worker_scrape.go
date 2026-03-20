// core/scraper/worker_scrape.go
package scraper

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
)

// runWorker is the main loop for a single agent goroutine. It bootstraps
// (browser + login + credits) then pulls queries from the channel.
func (s *Scraper) runWorker(
	ctx context.Context,
	agent *ScraperAgent,
	queries <-chan string,
	results chan<- queryResult,
	completed *atomic.Int32,
	empty *atomic.Int32,
	hostCount *atomic.Int32,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if !s.bootstrapWorker(ctx, agent) {
		return
	}
	logger.Debug().Str("agent", agent.Name).Msg("worker bootstrapped, entering query loop")

	defer s.shutdownWorker(agent)

	for {
		agent.SetStatusText(i18n.T("agent.waitingForQuery"))

		select {
		case <-ctx.Done():
			return
		case query, ok := <-queries:
			if !ok {
				agent.SetStatusText(i18n.T("agent.allQueriesDone"))
				return
			}

			result, _ := s.processQuery(ctx, agent, query)

			if result.Error != nil {
				// If no usable accounts remain, mark exhaustion and stop this worker.
				if errors.Is(result.Error, ErrNoUsableAccounts) {
					s.accountsExhausted.Store(true)
					logger.Warn().Str("agent", agent.Name).Msg("no usable accounts left, stopping worker")
					agent.SetStatusText(i18n.T("agent.noUsableAccounts"))
					s.recordQueryResult(result, completed, empty, hostCount)
					select {
					case results <- result:
					case <-ctx.Done():
					}
					return
				}

				// Reset agent to idle so it can attempt the next query.
				if agent.Status() == AgentStatusError {
					agent.SetStatus(AgentStatusIdle)
				}
			}

			s.recordQueryResult(result, completed, empty, hostCount)

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// bootstrapWorker launches the browser and logs in.
// Returns false if the worker should exit.
func (s *Scraper) bootstrapWorker(ctx context.Context, agent *ScraperAgent) bool {
	if ctx.Err() != nil {
		return false
	}

	if !s.launchBrowser(ctx, agent) {
		return false
	}

	if ctx.Err() != nil {
		return false
	}

	if err := s.bootstrapAccount(ctx, agent); err != nil {
		if ctx.Err() != nil {
			return false
		}
		if errors.Is(err, ErrNoUsableAccounts) {
			s.accountsExhausted.Store(true)
		}
		logger.Error().Err(err).Str("agent", agent.Name).Msg("bootstrap failed")
		agent.SetStatus(AgentStatusError)
		agent.SetStatusText(i18n.T("agent.bootstrapFailed"))
		s.emitEvent(EventAgentError, agent.Name, "bootstrap failed", nil)
		return false
	}

	agent.SetStatus(AgentStatusIdle)
	agent.SetStatusText(i18n.T("agent.ready"))
	s.emitEvent(EventAgentStarted, agent.Name, "worker ready", nil)
	logger.Info().Str("agent", agent.Name).Msg("worker ready")
	return true
}
