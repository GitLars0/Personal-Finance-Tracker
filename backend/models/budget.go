package models

import "time"

type Budget struct {
    ID          uint         `gorm:"primaryKey" json:"id"`
    UserID      uint         `gorm:"index;not null" json:"user_id"`
    User        User         `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
    PeriodStart time.Time    `gorm:"type:date;index;not null" json:"period_start"`
    PeriodEnd   time.Time    `gorm:"type:date;index;not null" json:"period_end"`
    Currency    string       `gorm:"size:3;not null;default:USD" json:"currency"`
    Items       []BudgetItem `json:"items,omitempty"`
    CreatedAt   time.Time    `json:"created_at"`
}

type BudgetItem struct {
    ID           uint     `gorm:"primaryKey" json:"id"`
    BudgetID     uint     `gorm:"index;not null" json:"budget_id"`
    CategoryID   uint     `gorm:"index;not null" json:"category_id"`
    Category     Category `json:"category,omitempty"`
    PlannedCents int64    `gorm:"not null" json:"planned_cents"`
}