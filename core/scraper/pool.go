// core/scraper/pool.go
package scraper

import (
	"sync"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// accountPool is a thread-safe pool of accounts for agent assignment.
// Each account is assigned to at most one agent at a time.
type accountPool struct {
	mu       sync.Mutex
	accounts map[uint]*models.Account // all accounts by ID
	assigned map[uint]string          // account ID -> agent name
}

func newAccountPool() *accountPool {
	return &accountPool{
		accounts: make(map[uint]*models.Account),
		assigned: make(map[uint]string),
	}
}

// loadFromDB populates the pool with all accounts from the database.
func (p *accountPool) loadFromDB() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	accounts, err := repositories.GetAccountRepository().ListAll()
	if err != nil {
		return err
	}

	for i := range accounts {
		p.accounts[accounts[i].ID] = &accounts[i]
	}

	logger.Info().Int("count", len(accounts)).Msg("account pool loaded from database")
	return nil
}

// assign picks the best unassigned account with enough credits.
// Priority: most credits, then most recent RefreshedCreditsAt.
// Returns nil if none have enough credits.
func (p *accountPool) assign(agentName string) *models.Account {
	p.mu.Lock()
	defer p.mu.Unlock()

	var best *models.Account
	for id, acc := range p.accounts {
		if _, taken := p.assigned[id]; taken {
			continue
		}
		if acc.CreditsAmount < CreditsAmountPerQuery {
			continue
		}
		if best == nil {
			best = acc
			continue
		}
		if acc.CreditsAmount > best.CreditsAmount {
			best = acc
		} else if acc.CreditsAmount == best.CreditsAmount {
			// Prefer the account refreshed more recently.
			accRefreshed := acc.RefreshedCreditsAt
			bestRefreshed := best.RefreshedCreditsAt
			if accRefreshed != nil && (bestRefreshed == nil || accRefreshed.After(*bestRefreshed)) {
				best = acc
			}
		}
	}

	if best != nil {
		p.assigned[best.ID] = agentName
		logger.Debug().
			Str("agent", agentName).
			Str("email", best.Email).
			Uint("credits", best.CreditsAmount).
			Msg("account assigned to agent")
	}

	return best
}

// usableCount returns the number of unassigned accounts with enough credits.
func (p *accountPool) usableCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	count := 0
	for id, acc := range p.accounts {
		if _, taken := p.assigned[id]; taken {
			continue
		}
		if acc.CreditsAmount >= CreditsAmountPerQuery {
			count++
		}
	}
	return count
}

// release unassigns the agent's current account, returning it to the pool.
func (p *accountPool) release(agentName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, name := range p.assigned {
		if name == agentName {
			delete(p.assigned, id)
			logger.Debug().Str("agent", agentName).Uint("account_id", id).Msg("account released from agent")
			return
		}
	}
}

// totalCount returns the total number of accounts in the pool.
func (p *accountPool) totalCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.accounts)
}

// fillRefreshChan loads accounts as refresh entries into the channel,
// skipping accounts whose credits have not been used since the last refresh.
func (p *accountPool) fillRefreshChan(ch chan<- refreshAccountEntry) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	skipped := 0
	for _, acc := range p.accounts {
		if acc.RefreshedCreditsAt != nil &&
			(acc.CreditsLastUsedAt == nil || acc.CreditsLastUsedAt.Before(*acc.RefreshedCreditsAt)) {
			logger.Debug().
				Str("email", acc.Email).
				Msg("skipping account refresh: credits not used since last refresh")
			skipped++
			continue
		}
		ch <- refreshAccountEntry{Account: acc, Retries: 0}
	}
	return skipped
}

// persistCredits writes the account's current credit amount back to the database.
func (p *accountPool) persistCredits(account *models.Account) {
	if err := repositories.GetAccountRepository().Update(account); err != nil {
		logger.Error().Err(err).Str("email", account.Email).Msg("failed to persist account credits")
	}
}
