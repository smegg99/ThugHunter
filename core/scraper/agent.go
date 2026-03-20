// core/scraper/agent.go
package scraper

import (
	"github.com/go-rod/rod"
	"github.com/smegg99/unrevealed"
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

// CreditsAmountPerQuery is the credit cost per search query.
const CreditsAmountPerQuery = 5

// AgentStatus represents the lifecycle state of a ScraperAgent.
type AgentStatus string

const (
	AgentStatusOffline AgentStatus = "offline" // not initialized or logged out
	AgentStatusWaiting AgentStatus = "waiting" // has browser, waiting for account
	AgentStatusIdle    AgentStatus = "idle"    // logged in, not busy
	AgentStatusBusy    AgentStatus = "busy"    // actively performing a task
	AgentStatusError   AgentStatus = "error"   // encountered a blocking error
)

// ScraperAgent wraps a headless browser session for scraping or account management.
type ScraperAgent struct {
	Name             string
	status           AgentStatus
	statusText       string               // human-readable current activity
	onStatusText     func(info AgentInfo) // callback for status text changes
	browser          *unrevealed.Browser
	account          *models.Account
	accountHint      string    // email hint shown before login completes
	page             *rod.Page // persistent single tab
	pageReady        bool      // true after first-time page setup
	estimatedCredits int       // local credit tracker to avoid extra network calls
}

// SetAccountHint stores the email of the account being attempted so the
// frontend can display it before login completes.
func (a *ScraperAgent) SetAccountHint(email string) {
	a.accountHint = email
}

// StatusText returns the human-readable description of the agent's current activity.
func (a *ScraperAgent) StatusText() string {
	return a.statusText
}

// SetStatusText updates the agent's activity description and fires the callback.
func (a *ScraperAgent) SetStatusText(text string) {
	a.statusText = text
	if a.onStatusText != nil {
		a.onStatusText(a.Info())
	}
}

// EstimatedCredits returns the locally tracked credit balance.
func (a *ScraperAgent) EstimatedCredits() int {
	return a.estimatedCredits
}

// DeductEstimatedCredits subtracts the per-query cost from the local estimate.
func (a *ScraperAgent) DeductEstimatedCredits() {
	a.estimatedCredits -= CreditsAmountPerQuery
	if a.estimatedCredits < 0 {
		a.estimatedCredits = 0
	}
	logger.Debug().Str("agent", a.Name).Int("estimated_credits", a.estimatedCredits).Msg("deducted estimated credits")
}

func (a *ScraperAgent) Status() AgentStatus {
	return a.status
}

func (a *ScraperAgent) SetStatus(status AgentStatus) {
	switch status {
	case AgentStatusIdle, AgentStatusBusy, AgentStatusOffline, AgentStatusWaiting:
		logger.Debug().Str("agent", a.Name).Str("status", string(status)).Msg("setting agent status")
	case AgentStatusError:
		logger.Error().Str("agent", a.Name).Str("status", string(status)).Msg("agent entered error status")
	default:
		logger.Warn().Str("agent", a.Name).Str("status", string(status)).Msg("attempted to set invalid agent status")
		return
	}
	a.status = status
}

// IsLoggedIn reports whether the agent has a valid account and is idle or busy.
func (a *ScraperAgent) IsLoggedIn() bool {
	return a.account != nil &&
		(a.status == AgentStatusBusy || a.status == AgentStatusIdle)
}

func (a *ScraperAgent) Account() *models.Account {
	return a.account
}

// Info returns a snapshot of the agent's current state.
func (a *ScraperAgent) Info() AgentInfo {
	info := AgentInfo{
		Name:             a.Name,
		Status:           a.status,
		StatusText:       a.statusText,
		EstimatedCredits: a.estimatedCredits,
	}
	if a.account != nil {
		info.Account = a.account.Email
		info.Credits = a.account.CreditsAmount
	} else if a.accountHint != "" {
		info.Account = a.accountHint
	}
	return info
}

func newScraperAgent(name string) *ScraperAgent {
	logger.Debug().Str("agent", name).Msg("creating agent")
	return &ScraperAgent{
		Name:   name,
		status: AgentStatusOffline,
	}
}
