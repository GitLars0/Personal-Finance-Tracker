package models

import "time"

type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	User        User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	AccountID   uint      `gorm:"index;not null" json:"account_id"`
	Account     Account   `json:"account,omitempty"`
	CategoryID  *uint     `gorm:"index" json:"category_id"`
	Category    *Category `json:"category,omitempty"`
	AmountCents int64     `gorm:"not null" json:"amount_cents"` // Amount in cents, +income, -expense
	Description string    `json:"description"`
	TxnDate     time.Time `gorm:"type:date;index;not null" json:"txn_date"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`

	// Bank integration fields
	BankTransactionID *string `gorm:"uniqueIndex" json:"bank_transaction_id,omitempty"`
	Metadata          *string `gorm:"type:text" json:"metadata,omitempty"` // Changed to string for SQLite compatibility

	Splits []TransactionSplit `gorm:"foreignKey:ParentTxnID" json:"splits,omitempty"`
}

type TransactionSplit struct {
	ID          uint     `gorm:"primaryKey" json:"id"`
	ParentTxnID uint     `gorm:"index;not null" json:"parent_txn_id"`
	CategoryID  uint     `gorm:"index;not null" json:"category_id"`
	Category    Category `json:"category,omitempty"`
	AmountCents int64    `gorm:"not null" json:"amount_cents"`
}
