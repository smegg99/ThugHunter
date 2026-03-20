// services/scraper/accounts.go
package scraper

import (
	"fmt"

	"smegg.me/thughunter/core/models"
	"smegg.me/thughunter/core/repositories"
)

// AccountPage holds a page of accounts plus total count.
type AccountPage struct {
	Items []models.Account `json:"items"`
	Total int64            `json:"total"`
}

// ListAccounts returns all persisted accounts.
func (s *Service) ListAccounts() ([]models.Account, error) {
	repo := repositories.GetAccountRepository()
	return repo.ListAll()
}

// ListAccountsPaginated returns a page of accounts sorted and optionally filtered server-side.
func (s *Service) ListAccountsPaginated(page, pageSize int, sortBy, sortOrder, search string) (*AccountPage, error) {
	repo := repositories.GetAccountRepository()
	items, total, err := repo.ListPaginated(page, pageSize, sortBy, sortOrder, search)
	if err != nil {
		return nil, err
	}
	return &AccountPage{Items: items, Total: total}, nil
}

// AccountCount returns the total number of persisted accounts.
func (s *Service) AccountCount() (int, error) {
	repo := repositories.GetAccountRepository()
	count, err := repo.Count()
	return int(count), err
}

// CanStartRun reports whether there are any accounts available to run with.
func (s *Service) CanStartRun() (bool, error) {
	count, err := s.AccountCount()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteAccount removes an account by ID.
func (s *Service) DeleteAccount(id uint) error {
	repo := repositories.GetAccountRepository()
	if err := repo.Delete(id); err != nil {
		return fmt.Errorf("delete account %d: %w", id, err)
	}
	emitServiceEvent(EventAccountsChanged, nil)
	return nil
}

// UpdateAccountPassword updates the password of an account.
func (s *Service) UpdateAccountPassword(id uint, password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	repo := repositories.GetAccountRepository()
	if err := repo.UpdatePassword(id, password); err != nil {
		return err
	}
	emitServiceEvent(EventAccountsChanged, nil)
	return nil
}

// AddAccount creates a user-defined account.
func (s *Service) AddAccount(email, password string) error {
	if email == "" || password == "" {
		return fmt.Errorf("email and password are required")
	}
	repo := repositories.GetAccountRepository()

	// Check for duplicate email, including soft-deleted records.
	exists, err := repo.EmailExists(email)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}
	if exists {
		return fmt.Errorf("account with email %q already exists", email)
	}

	account := models.NewAccount(email, password, "User", "Account", "Manual", true)
	if err := repo.Create(account); err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	emitServiceEvent(EventAccountsChanged, nil)
	return nil
}
