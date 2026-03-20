// core/scraper/service_persist.go
package scraper

import (
	"strconv"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// servicePersister saves one scraped service entry to its repository.
// host is the parent host (ID is set after upsert), port is the service port.
type servicePersister func(host *models.Host, port int) error

// persisterRegistry maps Censys service labels to their persister functions.
// Extend the system by calling RegisterServicePersister during init.
var persisterRegistry = map[string]servicePersister{
	"VNC": persistVNCService,
}

// RegisterServicePersister registers a persister for a given Censys service label.
func RegisterServicePersister(label string, p servicePersister) {
	persisterRegistry[label] = p
}

// storeServicesFromHosts saves known services from each host's Services map.
// Must be called after storeHosts so that host.ID values are populated.
func storeServicesFromHosts(hosts []*models.Host) {
	for _, host := range hosts {
		storeHostServices(host)
	}
}

// storeHostServices iterates a single host's Services map and dispatches to
// the appropriate persister for each recognised label.
func storeHostServices(host *models.Host) {
	if host.ID == 0 {
		// Host was not successfully saved; skip to avoid orphaned service rows.
		return
	}

	for label, ports := range host.Services {
		p, ok := persisterRegistry[label]
		if !ok {
			continue
		}

		for _, portStr := range ports {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				logger.Warn().
					Str("ip", host.IP).
					Str("label", label).
					Str("port", portStr).
					Msg("scraper: invalid port in services map, skipping")
				continue
			}

			if err := p(host, port); err != nil {
				logger.Warn().
					Err(err).
					Str("ip", host.IP).
					Int("port", port).
					Str("service", label).
					Msg("scraper: failed to persist service")
			}
		}
	}
}

// persistVNCService upserts a scraped VNC service stub by IP+port.
// Probe details (auth type, RFB version) are filled in later by the scanner
// and are preserved when an existing record is updated.
func persistVNCService(host *models.Host, port int) error {
	svc := &models.VNCService{
		ServiceBase: models.ServiceBase{
			HostID:      host.ID,
			IP:          host.IP,
			Port:        port,
			ServiceType: models.ServiceTypeVNC,
		},
		AuthType: models.VNCAuthUnknown,
	}
	return repositories.GetVNCServiceRepository().Upsert(svc, host.IP, port, host.ID)
}
