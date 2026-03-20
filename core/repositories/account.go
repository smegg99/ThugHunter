// core/repositories/account.go
package repositories

import (
	"fmt"
	"strings"

	"smegg.me/thughunter/common/logger"
	"smegg.me/thughunter/core/datastore"
	"smegg.me/thughunter/core/models"
)

// AccountRepository extends Repository with account-specific queries.
type AccountRepository struct {
	*Repository[models.Account]
}

// GetAccountRepository returns an AccountRepository backed by the global DB.
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

// EmailExists checks whether an email exists in the database, including soft-deleted records.
func (r *AccountRepository) EmailExists(email string) (bool, error) {
	db := datastore.Get()
	var count int64
	if err := db.Unscoped().Model(&models.Account{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check email exists %q: %w", email, err)
	}
	return count > 0, nil
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

// allowedSortColumns defines which columns can be sorted on to prevent injection.
var allowedSortColumns = map[string]bool{
	"email":                true,
	"credits_amount":       true,
	"credits_expire_at":    true,
	"refreshed_credits_at": true,
	"user_added":           true,
	"created_at":           true,
}

// ListPaginated returns a page of accounts with server-side sorting and optional search.
func (r *AccountRepository) ListPaginated(page, pageSize int, sortBy, sortOrder, search string) ([]models.Account, int64, error) {
	logger.Debug().Int("page", page).Int("pageSize", pageSize).Str("sortBy", sortBy).Str("sortOrder", sortOrder).Str("search", search).Msg("listing accounts paginated")

	db := datastore.Get()

	page, pageSize, offset := normalizePage(page, pageSize)

	base := db.Model(&models.Account{})
	if search = strings.TrimSpace(search); search != "" {
		like := "%" + search + "%"
		base = base.Where("email LIKE ?", like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count accounts: %w", err)
	}

	query := base
	if allowedSortColumns[sortBy] {
		order := "asc"
		if strings.EqualFold(sortOrder, "desc") {
			order = "desc"
		}
		if sortBy == "email" {
			query = query.Order("user_added DESC").Order("email " + order)
		} else {
			query = query.Order(sortBy + " " + order)
		}
	} else {
		query = query.Order("credits_amount desc")
	}

	var accounts []models.Account
	if err := query.Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		return nil, 0, fmt.Errorf("list accounts paginated: %w", err)
	}

	logger.Debug().Int("count", len(accounts)).Int64("total", total).Msg("accounts page loaded")
	return accounts, total, nil
}

// Count returns the total number of accounts.
func (r *AccountRepository) Count() (int64, error) {
	db := datastore.Get()
	var count int64
	if err := db.Model(&models.Account{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return count, nil
}

// TotalCredits returns the sum of all account credits.
func (r *AccountRepository) TotalCredits() (int64, error) {
	db := datastore.Get()
	var total int64
	if err := db.Model(&models.Account{}).Select("COALESCE(SUM(credits_amount), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("sum credits: %w", err)
	}
	return total, nil
}

// CountWithCredits returns the number of accounts that have credits > 0.
func (r *AccountRepository) CountWithCredits() (int64, error) {
	db := datastore.Get()
	var count int64
	if err := db.Model(&models.Account{}).Where("credits_amount > 0").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count accounts with credits: %w", err)
	}
	return count, nil
}

// UpdatePassword updates only the password field for an account.
func (r *AccountRepository) UpdatePassword(id uint, password string) error {
	logger.Debug().Uint("id", id).Msg("updating account password")
	db := datastore.Get()
	result := db.Model(&models.Account{}).Where("id = ?", id).Update("password", password)
	if result.Error != nil {
		return fmt.Errorf("update password for account %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account %d not found", id)
	}
	return nil
}
