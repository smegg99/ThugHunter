// services/scanner/hosts.go
package scanner

import (
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// HostPage holds a page of hosts with total count.
type HostPage struct {
	Items []models.Host `json:"items"`
	Total int64         `json:"total"`
}

// ListHosts returns a paginated, sorted list of hosts from the database.
// sortBy accepts: ip, city, country_code, ping_ms, hardware, created_at.
// Hosts with ping_ms = 0 (not yet scanned) are always sorted last.
func (s *Service) ListHosts(page, pageSize int, sortBy, sortOrder, search string, countries, labels []string, hardware string) (*HostPage, error) {
	repo := repositories.GetHostRepository()
	items, total, err := repo.ListPaginated(page, pageSize, sortBy, sortOrder, search, countries, labels, hardware)
	if err != nil {
		return nil, err
	}
	return &HostPage{Items: items, Total: total}, nil
}

// GetHostByIP returns a single host by its IP address.
func (s *Service) GetHostByIP(ip string) (*models.Host, error) {
	repo := repositories.GetHostRepository()
	return repo.FindByIP(ip)
}

// ToggleHostFavorite flips the is_favorite flag on a host and returns the new value.
func (s *Service) ToggleHostFavorite(id uint) (bool, error) {
	db := datastore.Get()
	var host models.Host
	if err := db.First(&host, id).Error; err != nil {
		return false, err
	}
	host.IsFavorite = !host.IsFavorite
	if err := db.Model(&host).Update("is_favorite", host.IsFavorite).Error; err != nil {
		return false, err
	}
	return host.IsFavorite, nil
}
