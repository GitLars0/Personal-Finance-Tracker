package models

import (
	"time"
	//"gorm.io/gorm"
)

// UserRole defines the possible roles for users
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // Never expose in JSON
	Name         string    `json:"name"`
	Role         UserRole  `gorm:"default:'user';not null" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
