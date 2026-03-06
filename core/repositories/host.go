package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

type HostRepository struct {
	*Repository[models.Host]
}

func NewHostRepository(db *gorm.DB) *HostRepository {
	return &HostRepository{Repository: New[models.Host](db)}
}

func (r *HostRepository) FindByIP(ip string) (*models.Host, error) {
	logger.Debug().Str("ip", ip).Msg("finding host by ip")

	var host models.Host
	if err := r.db.Where("ip = ?", ip).First(&host).Error; err != nil {
		return nil, fmt.Errorf("find host by ip %q: %w", ip, err)
	}
	return &host, nil
}

func (r *HostRepository) ListByCountry(countryCode string) ([]models.Host, error) {
	logger.Debug().Str("country", countryCode).Msg("listing hosts by country")

	var hosts []models.Host
	if err := r.db.Where("country_code = ?", countryCode).Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("list hosts by country %q: %w", countryCode, err)
	}

	logger.Debug().Str("country", countryCode).Int("count", len(hosts)).Msg("hosts found")
	return hosts, nil
}

func (r *HostRepository) ListByServiceType(serviceType models.ServiceType) ([]models.Host, error) {
	logger.Debug().Str("service_type", string(serviceType)).Msg("listing hosts by service type")

	var hosts []models.Host
	if err := r.db.Where("service_type = ?", serviceType).Find(&hosts).Error; err != nil {
		return nil, fmt.Errorf("list hosts by service type %q: %w", serviceType, err)
	}

	logger.Debug().Str("service_type", string(serviceType)).Int("count", len(hosts)).Msg("hosts found")
	return hosts, nil
}
