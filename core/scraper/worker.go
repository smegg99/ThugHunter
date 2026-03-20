// core/scraper/worker.go
package scraper

import (
	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// maxLoginAttempts is the maximum number of times to retry login that failed
// due to things like timeouts (turnstile might have not passed, the bots
// cannot click to pass it effectively from what I've observed) before giving up.
const maxLoginAttempts = 3

// maxRefreshRetries is the maximum number of times an account can be
// re-queued for refresh before it is removed from the queue.
const maxRefreshRetries = 2

// refreshAccountEntry wraps an account with a retry counter.
type refreshAccountEntry struct {
	Account *models.Account
	Retries int
}

// queryResult tracks the outcome of a single query execution.
type queryResult struct {
	Query string
	Hosts []*models.Host
	Error error
	Agent string
}

// storeHosts persists discovered hosts to the database, upserting by IP.
func storeHosts(hosts []*models.Host) {
	if len(hosts) == 0 {
		return
	}
	hostRepo := repositories.GetHostRepository()
	for _, host := range hosts {
		if err := hostRepo.Upsert(host); err != nil {
			logger.Warn().Err(err).Str("ip", host.IP).Msg("failed to upsert host")
		}
	}
}

// storeScrapedHosts persists discovered hosts and then saves the services
// extracted from each host's Services map to their respective repositories.
func storeScrapedHosts(hosts []*models.Host) {
	storeHosts(hosts)
	storeServicesFromHosts(hosts)
}
