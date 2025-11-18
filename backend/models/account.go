package models

import "time"

type AccountType string

const (
	AccountCash       AccountType = "cash"
	AccountChecking   AccountType = "checking"
	AccountSavings    AccountType = "savings"
	AccountCredit     AccountType = "credit"
	AccountInvestment AccountType = "investment"
	AccountOther      AccountType = "other"
)

type Account struct {
	ID                  uint        `gorm:"primaryKey" json:"id"`
	UserID              uint        `gorm:"index;not null" json:"user_id"`
	User                User        `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Name                string      `gorm:"not null" json:"name"`
	Type                AccountType `gorm:"type:text;not null" json:"account_type"`
	Currency            string      `gorm:"size:3;not null;default:USD" json:"currency"`
	InitialBalanceCents int64       `gorm:"default:0" json:"initial_balance_cents"`
	CurrentBalanceCents int64       `gorm:"default:0" json:"current_balance_cents"`
	Description         string      `json:"description"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}
