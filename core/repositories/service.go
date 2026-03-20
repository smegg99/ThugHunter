// core/repositories/service.go
package repositories

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
)

// ServiceRepository extends Repository with IP/port lookup methods.
type ServiceRepository[T any] struct {
	*Repository[T]
}

// NewServiceRepository returns a ServiceRepository for the given model type.
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

// InsertIfNotExists creates the entity only when no record with the same ip+port exists.
// This preserves any probe data already written by the scanner.
func (r *ServiceRepository[T]) InsertIfNotExists(entity *T, ip string, port int) error {
	db := datastore.Get()
	var existing T
	err := db.Where("ip = ? AND port = ?", ip, port).First(&existing).Error
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check service exists ip=%q port=%d: %w", ip, port, err)
	}
	return db.Create(entity).Error
}

// Upsert creates the entity if no record with the same ip+port exists, or
// updates the host_id of the existing record to keep it in sync with the
// parent host. Scanner-populated fields are preserved on update.
func (r *ServiceRepository[T]) Upsert(entity *T, ip string, port int, hostID uint) error {
	db := datastore.Get()
	var existing T
	err := db.Where("ip = ? AND port = ?", ip, port).First(&existing).Error
	if err == nil {
		return db.Model(&existing).Update("host_id", hostID).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("upsert service ip=%q port=%d: %w", ip, port, err)
	}
	if err := db.Create(entity).Error; err != nil {
		// Race condition: another goroutine inserted between our SELECT and INSERT.
		// Retry as an update.
		var retry T
		if findErr := db.Where("ip = ? AND port = ?", ip, port).First(&retry).Error; findErr == nil {
			return db.Model(&retry).Update("host_id", hostID).Error
		}
		return fmt.Errorf("upsert service ip=%q port=%d: %w", ip, port, err)
	}
	return nil
}
