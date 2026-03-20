// core/repositories/service_rdp.go
package repositories

import (
	"smegg.me/thughunter/core/models"
)

// RDPServiceRepository extends ServiceRepository for RDP services.
type RDPServiceRepository struct {
	*ServiceRepository[models.RDPService]
}

// GetRDPServiceRepository returns an RDPServiceRepository backed by the global DB.
func GetRDPServiceRepository() *RDPServiceRepository {
	return &RDPServiceRepository{ServiceRepository: NewServiceRepository[models.RDPService]()}
}
