// core/repositories/repository.go
package repositories

import (
	"fmt"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
)

// Repository is a generic CRUD wrapper around GORM.
type Repository[T any] struct {
}

func New[T any]() *Repository[T] {
	return &Repository[T]{}
}

func (r *Repository[T]) Create(entity *T) error {
	logger.Debug().Msg("creating entity")
	db := datastore.Get()
	return db.Create(entity).Error
}

func (r *Repository[T]) GetByID(id uint) (*T, error) {
	logger.Debug().Uint("id", id).Msg("getting entity by id")
	db := datastore.Get()

	var entity T
	if err := db.First(&entity, id).Error; err != nil {
		return nil, fmt.Errorf("get by id %d: %w", id, err)
	}
	return &entity, nil
}

func (r *Repository[T]) List() ([]T, error) {
	logger.Debug().Msg("listing entities")
	db := datastore.Get()

	var entities []T
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	logger.Debug().Int("count", len(entities)).Msg("entities listed")
	return entities, nil
}

func (r *Repository[T]) Update(entity *T) error {
	logger.Debug().Msg("updating entity")
	db := datastore.Get()
	return db.Save(entity).Error
}

func (r *Repository[T]) Delete(id uint) error {
	logger.Debug().Uint("id", id).Msg("deleting entity")
	db := datastore.Get()

	var entity T
	return db.Delete(&entity, id).Error
}

// normalizePage clamps page/pageSize to valid ranges and returns the offset.
func normalizePage(page, pageSize int) (normalizedPage, normalizedSize, offset int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize, (page - 1) * pageSize
}
