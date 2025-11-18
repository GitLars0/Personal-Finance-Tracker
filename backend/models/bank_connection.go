package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONB type for storing JSON data in PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

type BankConnection struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`

	// Bank Information
	BankName     string `json:"bank_name" gorm:"not null"`     // "sparebanken_norge" or "bulder_bank"
	BankEndpoint string `json:"bank_endpoint" gorm:"not null"` // https://psd2.spvapi.no or https://psd2-bulder.spvapi.no

	// PSD2 Consent Information
	ConsentID         string    `json:"consent_id" gorm:"uniqueIndex"`
	ConsentStatus     string    `json:"consent_status"` // received, valid, rejected, expired, etc.
	ConsentValidUntil time.Time `json:"consent_valid_until"`
	FrequencyPerDay   int       `json:"frequency_per_day" gorm:"default:4"`

	// Connection Status
	Status     string     `json:"status" gorm:"default:pending"` // pending, connected, expired, failed
	LastSyncAt *time.Time `json:"last_sync_at"`
	NextSyncAt *time.Time `json:"next_sync_at"`
	SyncCount  int        `json:"sync_count" gorm:"default:0"`

	// OAuth and sensitive data (for banks that use OAuth like Sparebank 1)
	Metadata JSONB `json:"metadata" gorm:"type:jsonb"`

	// Linked Accounts
	LinkedAccounts []BankAccount `json:"linked_accounts" gorm:"foreignKey:BankConnectionID"`
}

type BankAccount struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	BankConnectionID uint           `json:"bank_connection_id" gorm:"not null;index"`
	BankConnection   BankConnection `json:"-" gorm:"foreignKey:BankConnectionID"`

	// Bank Account Information
	AccountID   string `json:"account_id" gorm:"not null"` // The tokenized account ID from bank
	IBAN        string `json:"iban"`
	AccountName string `json:"account_name"`
	Currency    string `json:"currency" gorm:"default:NOK"`
	AccountType string `json:"account_type"` // checking, savings, card, etc.

	// Sync Information
	LastTransactionSync *time.Time `json:"last_transaction_sync"`
	IsActive            bool       `json:"is_active" gorm:"default:true"`

	// Link to internal account (optional)
	InternalAccountID *uint    `json:"internal_account_id"`
	InternalAccount   *Account `json:"internal_account,omitempty" gorm:"foreignKey:InternalAccountID"`
}

type BankSyncLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	BankConnectionID uint           `json:"bank_connection_id" gorm:"not null;index"`
	BankConnection   BankConnection `json:"-" gorm:"foreignKey:BankConnectionID"`

	SyncType          string `json:"sync_type"` // transactions, accounts, balances
	Status            string `json:"status"`    // success, failed, partial
	TransactionsFound int    `json:"transactions_found"`
	TransactionsAdded int    `json:"transactions_added"`
	ErrorMessage      string `json:"error_message,omitempty"`

	// API Usage tracking
	APICallsUsed int `json:"api_calls_used"`
	SyncDuration int `json:"sync_duration_ms"`
}
