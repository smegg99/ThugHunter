package repositories

import (
"gorm.io/gorm"

"smegg.me/thughunter/core/models"
)

type SPICEServiceRepository struct {
	*ServiceRepository[models.SPICEService]
}

func NewSPICEServiceRepository(db *gorm.DB) *SPICEServiceRepository {
	return &SPICEServiceRepository{ServiceRepository: NewServiceRepository[models.SPICEService](db)}
}
