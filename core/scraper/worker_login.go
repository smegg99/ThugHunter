// core/scraper/worker_login.go
package scraper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"smegg.me/thughunter/common/i18n"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

// loginWithRetry retries login up to maxLoginAttempts times.
func (s *Scraper) loginWithRetry(ctx context.Context, agent *ScraperAgent, account *models.Account) error {
	agent.SetAccountHint(account.Email)

	isNotActive := func(err error) bool { return errors.Is(err, ErrAccountNotActive) }
	isTimeout := func(err error) bool { return errors.Is(err, ErrTimeout) }

	for attempt := 1; attempt <= maxLoginAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		agent.SetStatus(AgentStatusBusy)
		agent.SetStatusText(i18n.T("agent.loggingIn"))
		err := agent.Login(ctx, account)
		if err == nil {
			agent.SetStatusText(i18n.T("agent.loggedIn"))
			return nil
		}

		if isNotActive(err) {
			logger.Warn().Str("agent", agent.Name).Str("email", account.Email).Msg("account not active, discarding")
			agent.SetStatus(AgentStatusError)
			agent.SetStatusText(i18n.T("agent.accountNotActive"))
			return fmt.Errorf("login as %s: %w", account.Email, ErrAccountNotActive)
		}

		if isTimeout(err) {
			logger.Warn().Str("agent", agent.Name).Str("email", account.Email).Int("attempt", attempt).Msg("login timed out, retrying")
			agent.SetStatusText(i18n.Tf("agent.loginTimedOut", attempt, maxLoginAttempts))
			continue
		}

		// Non-retryable error.
		agent.SetStatus(AgentStatusError)
		agent.SetStatusText(i18n.T("agent.loginFailed"))
		return err
	}

	agent.SetStatus(AgentStatusError)
	agent.SetStatusText(i18n.Tf("agent.loginExhausted", maxLoginAttempts))
	return fmt.Errorf("login as %s: exhausted %d attempts", account.Email, maxLoginAttempts)
}

// bootstrapAccount tries pool accounts until one succeeds.
// Returns ErrNoUsableAccounts if no pool account could be logged in.
func (s *Scraper) bootstrapAccount(ctx context.Context, agent *ScraperAgent) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	for {
		account := s.pool.assign(agent.Name)
		if account == nil {
			logger.Warn().Str("agent", agent.Name).Msg("no usable accounts left in pool")
			agent.SetStatus(AgentStatusError)
			agent.SetStatusText(i18n.T("agent.noUsableAccounts"))
			return ErrNoUsableAccounts
		}
		if ctx.Err() != nil {
			s.pool.release(agent.Name)
			return ctx.Err()
		}
		if err := s.loginWithRetry(ctx, agent, account); err != nil {
			logger.Warn().Str("agent", agent.Name).Str("email", account.Email).Msg("pool login failed, trying next account")
			s.pool.release(agent.Name)
			agent.SetAccountHint("")
			_ = agent.ClearSession()
			continue
		}

		// Always verify real credits after login - DB values may be stale.
		logger.Debug().Str("agent", agent.Name).Str("email", account.Email).Msg("login succeeded, verifying credits")
		if err := s.bootstrapCredits(ctx, agent); err != nil {
			logger.Warn().Str("agent", agent.Name).Str("email", account.Email).Msg("credit check failed after login, trying next account")
			s.pool.release(agent.Name)
			agent.SetAccountHint("")
			_ = agent.ClearSession()
			continue
		}

		// If the real balance is too low, skip this account.
		if agent.EstimatedCredits() < CreditsAmountPerQuery {
			logger.Warn().Str("agent", agent.Name).Str("email", account.Email).
				Int("credits", agent.EstimatedCredits()).
				Msg("account has insufficient credits after verification, trying next")
			s.pool.release(agent.Name)
			agent.SetAccountHint("")
			_ = agent.ClearSession()
			continue
		}

		return nil
	}
}

// bootstrapCredits fetches real credits after a successful login.
// Returns the underlying error if the credit fetch fails.
// Context cancellation is returned silently without logging.
func (s *Scraper) bootstrapCredits(ctx context.Context, agent *ScraperAgent) error {
	agent.SetStatus(AgentStatusBusy)
	agent.SetStatusText(i18n.T("agent.checkingCredits"))
	credits, err := agent.UpdateCredits(ctx)
	if err != nil {
		if ctx.Err() != nil {
			agent.SetStatus(AgentStatusIdle)
			return ctx.Err()
		}
		agent.SetStatus(AgentStatusError)
		agent.SetStatusText(i18n.T("agent.creditUpdateFailed"))
		logger.Warn().Err(err).Str("agent", agent.Name).Msg("failed to update credits during bootstrap")
		return err
	}

	agent.SetStatusText(i18n.T("agent.updatedCredits"))

	persistCredits := func() {
		account := agent.Account()
		if account == nil {
			return
		}
		account.CreditsAmount = uint(credits.Current)
		account.CreditsExpireAt = credits.ExpiresAt
		if uint(credits.Current) < CreditsAmountPerQuery {
			agent.SetStatusText(i18n.T("agent.ranOutOfCredits"))
			now := time.Now()
			account.RanOutOfCreditsAt = &now
		}
		s.pool.persistCredits(account)
	}

	persistCredits()

	s.emitEvent(EventCreditsRefreshed, agent.Name, fmt.Sprintf("credits: %d", credits.Current), map[string]any{
		"credits": credits.Current,
	})

	agent.SetStatus(AgentStatusIdle)
	return nil
}
