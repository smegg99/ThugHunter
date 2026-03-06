package repositories

import (
	"fmt"

	"gorm.io/gorm"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/models"
)

type AccountRepository struct {
	*Repository[models.Account]
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{Repository: New[models.Account](db)}
}

func (r *AccountRepository) FindByEmail(email string) (*models.Account, error) {
	logger.Debug().Str("email", email).Msg("finding account by email")

	var account models.Account
	if err := r.db.Where("email = ?", email).First(&account).Error; err != nil {
		return nil, fmt.Errorf("find account by email %q: %w", email, err)
	}
	return &account, nil
}

func (r *AccountRepository) ListWithCredits() ([]models.Account, error) {
	logger.Debug().Msg("listing accounts with credits")

	var accounts []models.Account
	if err := r.db.Where("credits_amount > 0").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list accounts with credits: %w", err)
	}

	logger.Debug().Int("count", len(accounts)).Msg("accounts with credits found")
	return accounts, nil
}
