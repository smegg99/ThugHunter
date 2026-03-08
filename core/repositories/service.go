package repositories

import (
	"fmt"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
)

type ServiceRepository[T any] struct {
	*Repository[T]
}

func NewServiceRepository[T any]() *ServiceRepository[T] {
	return &ServiceRepository[T]{Repository: New[T]()}
}

func (r *ServiceRepository[T]) FindByIP(ip string) ([]T, error) {
	logger.Debug().Str("ip", ip).Msg("finding services by ip")

	db := datastore.Get()
	var services []T
	if err := db.Where("ip = ?", ip).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("find services by ip %q: %w", ip, err)
	}

	logger.Debug().Str("ip", ip).Int("count", len(services)).Msg("services found")
	return services, nil
}

func (r *ServiceRepository[T]) FindByPort(port int) ([]T, error) {
	logger.Debug().Int("port", port).Msg("finding services by port")

	db := datastore.Get()
	var services []T
	if err := db.Where("port = ?", port).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("find services by port %d: %w", port, err)
	}

	logger.Debug().Int("port", port).Int("count", len(services)).Msg("services found")
	return services, nil
}

func (r *ServiceRepository[T]) FindByIPAndPort(ip string, port int) (*T, error) {
	logger.Debug().Str("ip", ip).Int("port", port).Msg("finding service by ip and port")

	db := datastore.Get()
	var service T
	if err := db.Where("ip = ? AND port = ?", ip, port).First(&service).Error; err != nil {
		return nil, fmt.Errorf("find service by ip %q port %d: %w", ip, port, err)
	}
	return &service, nil
}
