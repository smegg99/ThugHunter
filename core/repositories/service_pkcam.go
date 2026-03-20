// core/repositories/service_pkcam.go
package repositories

import (
	"fmt"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

// PKCameraServiceRepository extends ServiceRepository for P2P/K cameras.
type PKCameraServiceRepository struct {
	*ServiceRepository[models.PKCameraService]
}

// GetPKCameraServiceRepository returns a PKCameraServiceRepository backed by the global DB.
func GetPKCameraServiceRepository() *PKCameraServiceRepository {
	return &PKCameraServiceRepository{ServiceRepository: NewServiceRepository[models.PKCameraService]()}
}

func (r *PKCameraServiceRepository) ListWithTwoWayAudio() ([]models.PKCameraService, error) {
	logger.Debug().Msg("listing pkcamera services with two-way audio")

	db := datastore.Get()
	var services []models.PKCameraService
	if err := db.Where("has_two_way_audio = ?", true).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("list pkcamera with two-way audio: %w", err)
	}

	logger.Debug().Int("count", len(services)).Msg("pkcamera services with two-way audio found")
	return services, nil
}
