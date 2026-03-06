package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
)

type ServiceRepository[T any] struct {
	*Repository[T]
}

func NewServiceRepository[T any](db *gorm.DB) *ServiceRepository[T] {
	return &ServiceRepository[T]{Repository: New[T](db)}
}

func (r *ServiceRepository[T]) FindByIP(ip string) ([]T, error) {
	logger.Debug().Str("ip", ip).Msg("finding services by ip")

	var services []T
	if err := r.db.Where("ip = ?", ip).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("find services by ip %q: %w", ip, err)
	}

	logger.Debug().Str("ip", ip).Int("count", len(services)).Msg("services found")
	return services, nil
}

func (r *ServiceRepository[T]) FindByPort(port int) ([]T, error) {
	logger.Debug().Int("port", port).Msg("finding services by port")

	var services []T
	if err := r.db.Where("port = ?", port).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("find services by port %d: %w", port, err)
	}

	logger.Debug().Int("port", port).Int("count", len(services)).Msg("services found")
	return services, nil
}

func (r *ServiceRepository[T]) FindByIPAndPort(ip string, port int) (*T, error) {
	logger.Debug().Str("ip", ip).Int("port", port).Msg("finding service by ip and port")

	var service T
	if err := r.db.Where("ip = ? AND port = ?", ip, port).First(&service).Error; err != nil {
		return nil, fmt.Errorf("find service by ip %q port %d: %w", ip, port, err)
	}
	return &service, nil
}
