// services/scraper/dashboard.go
package scraper

import (
	"smegg.me/thughunter/core/repositories"
)

// DashboardStats holds aggregate statistics for the dashboard.
type DashboardStats struct {
	TotalHosts     int64 `json:"total_hosts"`
	TotalAccounts  int64 `json:"total_accounts"`
	TotalCredits   int64 `json:"total_credits"`
	UsableAccounts int64 `json:"usable_accounts"`
}

// GetDashboardStats returns aggregate stats for the home page dashboard.
func (s *Service) GetDashboardStats() (*DashboardStats, error) {
	accountRepo := repositories.GetAccountRepository()
	hostRepo := repositories.GetHostRepository()

	accounts, err := accountRepo.Count()
	if err != nil {
		return nil, err
	}

	credits, err := accountRepo.TotalCredits()
	if err != nil {
		return nil, err
	}

	hosts, err := hostRepo.Count()
	if err != nil {
		return nil, err
	}

	usable, err := accountRepo.CountWithCredits()
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalHosts:     hosts,
		TotalAccounts:  accounts,
		TotalCredits:   credits,
		UsableAccounts: usable,
	}, nil
}
