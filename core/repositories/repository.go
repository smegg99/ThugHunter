package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
)

type Repository[T any] struct {
	db *gorm.DB
}

func New[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (r *Repository[T]) Create(entity *T) error {
	logger.Debug().Msg("creating entity")
	return r.db.Create(entity).Error
}

func (r *Repository[T]) GetByID(id uint) (*T, error) {
	logger.Debug().Uint("id", id).Msg("getting entity by id")

	var entity T
	if err := r.db.First(&entity, id).Error; err != nil {
		return nil, fmt.Errorf("get by id %d: %w", id, err)
	}
	return &entity, nil
}

func (r *Repository[T]) List() ([]T, error) {
	logger.Debug().Msg("listing entities")

	var entities []T
	if err := r.db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	logger.Debug().Int("count", len(entities)).Msg("entities listed")
	return entities, nil
}

func (r *Repository[T]) Update(entity *T) error {
	logger.Debug().Msg("updating entity")
	return r.db.Save(entity).Error
}

func (r *Repository[T]) Delete(id uint) error {
	logger.Debug().Uint("id", id).Msg("deleting entity")

	var entity T
	return r.db.Delete(&entity, id).Error
}
