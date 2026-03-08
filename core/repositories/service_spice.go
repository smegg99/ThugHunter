package repositories

import (
	"smegg.me/thughunter/core/models"
)

type SPICEServiceRepository struct {
	*ServiceRepository[models.SPICEService]
}

func GetSPICEServiceRepository() *SPICEServiceRepository {
	return &SPICEServiceRepository{ServiceRepository: NewServiceRepository[models.SPICEService]()}
}
