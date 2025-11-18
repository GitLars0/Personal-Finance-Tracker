package models

import "time"

type CategoryKind string

const (
	CategoryExpense CategoryKind = "expense"
	CategoryIncome  CategoryKind = "income"
)

type Category struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	UserID      uint         `gorm:"index;not null" json:"user_id"`
	User        User         `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Name        string       `gorm:"not null" json:"name"`
	ParentID    *uint        `gorm:"index" json:"parent_id"`
	Kind        CategoryKind `gorm:"type:text;not null" json:"kind"`
	Description *string      `gorm:"type:text" json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
}
