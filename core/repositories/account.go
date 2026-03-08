package repositories

import (
	"fmt"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

type AccountRepository struct {
	*Repository[models.Account]
}

func GetAccountRepository() *AccountRepository {
	return &AccountRepository{Repository: New[models.Account]()}
}

func (r *AccountRepository) FindByEmail(email string) (*models.Account, error) {
	logger.Debug().Str("email", email).Msg("finding account by email")

	db := datastore.Get()
	var account models.Account
	if err := db.Where("email = ?", email).First(&account).Error; err != nil {
		return nil, fmt.Errorf("find account by email %q: %w", email, err)
	}
	return &account, nil
}

func (r *AccountRepository) ListWithCredits() ([]models.Account, error) {
	logger.Debug().Msg("listing accounts with credits")

	db := datastore.Get()
	var accounts []models.Account
	if err := db.Where("credits_amount > 0").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list accounts with credits: %w", err)
	}

	logger.Debug().Int("count", len(accounts)).Msg("accounts with credits found")
	return accounts, nil
}

func (r *AccountRepository) ListAll() ([]models.Account, error) {
	logger.Debug().Msg("listing all accounts")

	db := datastore.Get()
	var accounts []models.Account
	if err := db.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list all accounts: %w", err)
	}

	logger.Debug().Int("count", len(accounts)).Msg("accounts found")
	return accounts, nil
}