package repositories

import (
"gorm.io/gorm"

"smegg.me/thughunter/core/models"
)

type RDPServiceRepository struct {
	*ServiceRepository[models.RDPService]
}

func NewRDPServiceRepository(db *gorm.DB) *RDPServiceRepository {
	return &RDPServiceRepository{ServiceRepository: NewServiceRepository[models.RDPService](db)}
}
