// core/scraper/worker_register.go
package scraper

import (
	"context"
	"fmt"
	"sync"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
)

// runRegisterWorker is the per-agent goroutine for register mode.
// It launches a browser and registers accounts in a loop until the target
// count is reached, the context is cancelled, or a duration timeout fires.
func (s *Scraper) runRegisterWorker(
	ctx context.Context,
	agent *ScraperAgent,
	results chan<- registerResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if !s.launchBrowser(ctx, agent) {
		return
	}

	s.emitEvent(EventAgentStarted, agent.Name, "register worker ready", nil)
	defer s.shutdownWorker(agent)

	target := int(s.targetAccounts.Load())

	for {
		if ctx.Err() != nil {
			return
		}

		// Check if we've hit the target (0 = unlimited).
		if target > 0 && int(s.createdAccounts.Load()) >= target {
			agent.SetStatusText(i18n.T("agent.targetReached"))
			return
		}

		agent.SetStatus(AgentStatusBusy)
		agent.SetStatusText(i18n.T("agent.registeringAccount"))

		// Clear the previous session before starting a new registration.
		_ = agent.ClearSession()

		account, err := agent.Register(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			if isConnectionError(err) {
				if !s.relaunchOnConnectionError(ctx, agent, err) {
				}
			}

			logger.Warn().Err(err).Str("agent", agent.Name).Msg("registration failed")
			agent.SetStatus(AgentStatusError)
			agent.SetStatusText(i18n.T("agent.registrationFailed"))
			s.emitEvent(EventAgentError, agent.Name, "registration failed", map[string]any{
				"error": err.Error(),
			})

			select {
			case results <- registerResult{Error: err, Agent: agent.Name}:
			case <-ctx.Done():
				return
			}

			continue
		}

		agent.SetStatus(AgentStatusIdle)
		agent.SetAccountHint(account.Email)
		logger.Info().Str("agent", agent.Name).Str("email", account.Email).Msg("account registered")
		agent.SetStatusText(i18n.T("agent.registered"))
		s.emitEvent(EventAccountCreated, agent.Name, fmt.Sprintf("registered %s", account.Email), map[string]any{
			"email": account.Email,
		})

		select {
		case results <- registerResult{Email: account.Email, Agent: agent.Name}:
		case <-ctx.Done():
			return
		}
	}
}
