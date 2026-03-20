// core/repositories/host.go
package repositories

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

// HostRepository extends Repository with host-specific queries.
type HostRepository struct {
	*Repository[models.Host]
}

// GetHostRepository returns a HostRepository backed by the global DB.
func GetHostRepository() *HostRepository {
	return &HostRepository{Repository: New[models.Host]()}
}

// Upsert inserts a new host or updates the existing one if a host with the same
// IP already exists. All mutable fields are overwritten with the new values.
func (r *HostRepository) Upsert(host *models.Host) error {
	db := datastore.Get()
	var existing models.Host
	if err := db.Where("ip = ?", host.IP).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(host).Error
		}
		return fmt.Errorf("upsert host %q: %w", host.IP, err)
	}

	host.ID = existing.ID
	host.CreatedAt = existing.CreatedAt
	host.IsFavorite = existing.IsFavorite
	return db.Save(host).Error
}

func (r *HostRepository) FindByIP(ip string) (*models.Host, error) {
	logger.Debug().Str("ip", ip).Msg("finding host by ip")

	db := datastore.Get()
	var host models.Host
	if err := db.Where("ip = ?", ip).First(&host).Error; err != nil {
		return nil, fmt.Errorf("find host by ip %q: %w", ip, err)
	}
	return &host, nil
}

func (r *HostRepository) ListByCountry(countryCode string) ([]models.Host, error) {
	logger.Debug().Str("country", countryCode).Msg("listing hosts by country")

	db := datastore.Get()
	var hosts []models.Host
	if err := db.Where("country_code = ?", countryCode).Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("list hosts by country %q: %w", countryCode, err)
	}

	logger.Debug().Str("country", countryCode).Int("count", len(hosts)).Msg("hosts found")
	return hosts, nil
}

// Count returns the total number of hosts.
func (r *HostRepository) Count() (int64, error) {
	db := datastore.Get()
	var count int64
	if err := db.Model(&models.Host{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count hosts: %w", err)
	}
	return count, nil
}

// allowedHostSortCols is the set of columns that may be used for sorting.
var allowedHostSortCols = map[string]bool{
	"ip":           true,
	"city":         true,
	"country_code": true,
	"ping_ms":      true,
	"created_at":   true,
}

// ListPaginated returns a page of hosts with optional search, sorting and filters.
// Hosts with ping_ms = 0 (not yet scanned) are sorted last.
func (r *HostRepository) ListPaginated(page, pageSize int, sortBy, sortOrder, search string, countries, labels []string, hardware string) ([]models.Host, int64, error) {
	db := datastore.Get()

	page, pageSize, offset := normalizePage(page, pageSize)

	base := db.Model(&models.Host{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		base = base.Where("ip LIKE ? OR city LIKE ? OR country_code LIKE ?", like, like, like)
	}
	if len(countries) > 0 {
		base = base.Where("country_code IN ?", countries)
	}
	if len(labels) > 0 {
		for _, lbl := range labels {
			like := "%" + lbl + "%"
			base = base.Where("labels LIKE ?", like)
		}
	}
	if strings.TrimSpace(hardware) != "" {
		base = base.Where("hardware LIKE ?", "%"+strings.TrimSpace(hardware)+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count hosts: %w", err)
	}

	col := "ping_ms"
	if allowedHostSortCols[sortBy] {
		col = sortBy
	}
	ord := "asc"
	if strings.ToLower(sortOrder) == "desc" {
		ord = "desc"
	}
	// Put unscanned hosts (ping_ms = 0 or NULL) at the end, always sort by ping first.
	// Favorites always come first regardless of sort.
	favFirst := `CASE WHEN COALESCE(is_favorite, 0) = 1 THEN 0 ELSE 1 END`
	var orderExpr string
	if col == "ping_ms" {
		orderExpr = fmt.Sprintf("%s, CASE WHEN ping_ms IS NULL OR ping_ms = 0 THEN 1 ELSE 0 END, ping_ms %s", favFirst, ord)
	} else {
		orderExpr = fmt.Sprintf("%s, CASE WHEN ping_ms IS NULL OR ping_ms = 0 THEN 1 ELSE 0 END, ping_ms ASC, %s %s", favFirst, col, ord)
	}

	hosts := make([]models.Host, 0)
	if err := base.Order(orderExpr).Offset(offset).Limit(pageSize).Find(&hosts).Error; err != nil {
		return nil, 0, fmt.Errorf("list hosts paginated: %w", err)
	}
	return hosts, total, nil
}
