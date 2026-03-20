// services/scanner/stats.go
package scanner

import (
	"fmt"

	"smegg.me/thughunter/core/datastore"
)

// BrowseStats contains aggregate statistics for the browse UI.
type BrowseStats struct {
	TotalHosts    int64 `json:"total_hosts"`
	PingOKHosts   int64 `json:"ping_ok_hosts"`
	TotalVNC      int64 `json:"total_vnc"`
	NoAuthVNC     int64 `json:"no_auth_vnc"`
	ScreenshotVNC int64 `json:"screenshot_vnc"`
}

// GetBrowseStats returns aggregate statistics for the browse page.
func (s *Service) GetBrowseStats() (*BrowseStats, error) {
	db := datastore.Get()
	stats := &BrowseStats{}

	if err := db.Table("hosts").Count(&stats.TotalHosts).Error; err != nil {
		return nil, fmt.Errorf("count hosts: %w", err)
	}
	if err := db.Table("hosts").Where("ping_ms > 0").Count(&stats.PingOKHosts).Error; err != nil {
		return nil, fmt.Errorf("count ping ok hosts: %w", err)
	}
	if err := db.Table("vnc_services").Count(&stats.TotalVNC).Error; err != nil {
		return nil, fmt.Errorf("count vnc services: %w", err)
	}
	if err := db.Table("vnc_services").Where("no_auth = 1").Count(&stats.NoAuthVNC).Error; err != nil {
		return nil, fmt.Errorf("count no auth vnc: %w", err)
	}
	if err := db.Table("vnc_services").Where("length(screenshot) > 0").Count(&stats.ScreenshotVNC).Error; err != nil {
		return nil, fmt.Errorf("count screenshot vnc: %w", err)
	}

	return stats, nil
}
