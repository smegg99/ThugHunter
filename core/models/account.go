// core/models/crawler_account.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Email             string         `gorm:"uniqueIndex;not null" json:"email"`
	Password          string         `gorm:"not null" json:"password"`
	FirstName         string         `gorm:"not null" json:"first_name"`
	LastName          string         `gorm:"not null" json:"last_name"`
	Organization      string         `gorm:"not null" json:"organization"`
	CreditsAmount     uint           `gorm:"default:0" json:"credits_amount"`
	RanOutOfCreditsAt *time.Time     `json:"ran_out_of_credits_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func NewAccount(email, password, firstName, lastName, organization string) *Account {
	return &Account{
		Email:        email,
		Password:     password,
		FirstName:    firstName,
		LastName:     lastName,
		Organization: organization,
	}
}
