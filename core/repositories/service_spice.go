// core/repositories/service_spice.go
package repositories

import (
	"smegg.me/thughunter/core/models"
)

// SPICEServiceRepository extends ServiceRepository for SPICE services.
type SPICEServiceRepository struct {
	*ServiceRepository[models.SPICEService]
}

// GetSPICEServiceRepository returns a SPICEServiceRepository backed by the global DB.
func GetSPICEServiceRepository() *SPICEServiceRepository {
	return &SPICEServiceRepository{ServiceRepository: NewServiceRepository[models.SPICEService]()}
}
