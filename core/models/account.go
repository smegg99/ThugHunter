// core/models/account.go
//
// Account model for Censys search accounts used by the scraper.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Account stores credentials and credit state for a Censys search account.
type Account struct {
	gorm.Model
	Email              string     `gorm:"uniqueIndex;not null" json:"email"`
	Password           string     `gorm:"not null" json:"password"`
	FirstName          string     `gorm:"not null" json:"first_name"`
	LastName           string     `gorm:"not null" json:"last_name"`
	Organization       string     `gorm:"not null" json:"organization"`
	CreditsAmount      uint       `gorm:"default:0" json:"credits_amount"`
	UserAdded          bool       `gorm:"default:false" json:"user_added"` // Users can add accounts manually, but we want to distinguish them from those added via config for better UX and error handling.
	CreditsLastUsedAt  *time.Time `json:"credits_last_used_at,omitempty"`  // Track when credits were last used to help identify accounts that do not need to be refreshed.
	CreditsExpireAt    *time.Time `json:"credits_expire_at,omitempty"`
	RanOutOfCreditsAt  *time.Time `json:"ran_out_of_credits_at,omitempty"`
	RefreshedCreditsAt *time.Time `json:"refreshed_credits_at,omitempty"`
}

// NewAccount creates an Account with the required fields.
func NewAccount(email, password, firstName, lastName, organization string, userAdded bool) *Account {
	return &Account{
		Email:        email,
		Password:     password,
		FirstName:    firstName,
		LastName:     lastName,
		Organization: organization,
		UserAdded:    userAdded,
	}
}

func (a *Account) IsUserAdded() bool {
	return a.UserAdded
}

func (a *Account) HasCredits() bool {
	return a.CreditsAmount > 0
}

// CreditsExpired reports whether the credit expiration date has passed.
// Returns true if no expiration date is set (unknown = treat as expired).
func (a *Account) CreditsExpired() bool {
	if a.CreditsExpireAt == nil {
		return true
	}
	return time.Now().After(*a.CreditsExpireAt)
}

func (a *Account) IsValid() bool {
	return a.Email != "" && a.Password != "" && a.FirstName != "" && a.LastName != "" && a.Organization != ""
}
