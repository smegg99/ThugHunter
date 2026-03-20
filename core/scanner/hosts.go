// core/scanner/hosts.go
package scanner

import (
	"fmt"

	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// loadHosts returns all hosts from the database.
func loadHosts() ([]*models.Host, error) {
	hosts, err := repositories.GetHostRepository().List()
	if err != nil {
		return nil, fmt.Errorf("load hosts: %w", err)
	}

	ptrs := make([]*models.Host, len(hosts))
	for i := range hosts {
		ptrs[i] = &hosts[i]
	}
	return ptrs, nil
}
