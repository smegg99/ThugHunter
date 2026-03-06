package repositories

import (
"gorm.io/gorm"

"smegg.me/thughunter/core/models"
)

type VNCServiceRepository struct {
	*ServiceRepository[models.VNCService]
}

func NewVNCServiceRepository(db *gorm.DB) *VNCServiceRepository {
	return &VNCServiceRepository{ServiceRepository: NewServiceRepository[models.VNCService](db)}
}
