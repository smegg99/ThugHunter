package repositories

import (
	"smegg.me/thughunter/core/models"
)

type RDPServiceRepository struct {
	*ServiceRepository[models.RDPService]
}

func GetRDPServiceRepository() *RDPServiceRepository {
	return &RDPServiceRepository{ServiceRepository: NewServiceRepository[models.RDPService]()}
}
