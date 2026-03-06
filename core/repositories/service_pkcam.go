package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

type PKCameraServiceRepository struct {
	*ServiceRepository[models.PKCameraService]
}

func NewPKCameraServiceRepository(db *gorm.DB) *PKCameraServiceRepository {
	return &PKCameraServiceRepository{ServiceRepository: NewServiceRepository[models.PKCameraService](db)}
}

func (r *PKCameraServiceRepository) ListWithTwoWayAudio() ([]models.PKCameraService, error) {
	logger.Debug().Msg("listing pkcamera services with two-way audio")

	var services []models.PKCameraService
	if err := r.db.Where("has_two_way_audio = ?", true).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("list pkcamera with two-way audio: %w", err)
	}

	logger.Debug().Int("count", len(services)).Msg("pkcamera services with two-way audio found")
	return services, nil
}
