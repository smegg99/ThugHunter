package repositories

import (
	"smegg.me/thughunter/core/models"
)

type VNCServiceRepository struct {
	*ServiceRepository[models.VNCService]
}

func GetNCServiceRepository() *VNCServiceRepository {
	return &VNCServiceRepository{ServiceRepository: NewServiceRepository[models.VNCService]()}
}
