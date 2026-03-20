// core/scraper/errors.go
package scraper

import "errors"

var (
	// Agent lifecycle errors
	ErrAgentNotLoggedIn     = errors.New("agent is not logged in")
	ErrNoAccountProvided    = errors.New("no account provided")
	ErrMaxAgentsReached     = errors.New("max agents reached")
	ErrAgentNotFound        = errors.New("agent not found")
	ErrNameGenerationFailed = errors.New("failed to generate unique agent name")

	// Browser errors
	ErrBrowserLaunchFailed = errors.New("browser launch failed")
	ErrNavigationFailed    = errors.New("navigation failed")
	ErrElementNotFound     = errors.New("element not found")

	// Request / response errors
	ErrInvalidURL           = errors.New("invalid URL")
	ErrTimeout              = errors.New("timed out")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrResponseReadFailed   = errors.New("failed to read response body")

	// Content errors
	ErrParseFailed     = errors.New("failed to parse content")
	ErrNoContentFound  = errors.New("no content found")
	ErrRanOutOfCredits = errors.New("ran out of credits")

	// Workflow errors
	ErrLoginFailed        = errors.New("login failed")
	ErrAccountNotActive   = errors.New("account not active")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRegistrationFailed = errors.New("registration failed")
	ErrVerificationFailed = errors.New("verification failed")
	ErrScrapeFailed       = errors.New("scrape failed")

	// Run errors
	ErrScraperAlreadyRunning = errors.New("scraper is already running")
	ErrNoQueries             = errors.New("no queries to process")
	ErrNoUsableAccounts      = errors.New("no usable accounts available")
)
