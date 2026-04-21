package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        int64          `gorm:"primaryKey"`
	UserID    int64          `json:"user_id"`
	Email     string         `json:"email"`
	Token     string         `json:"token"`
	Status    bool           `json:"status" gorm:"default:true"`
	ExpiresAt time.Time      `json:"expires_at"`
	UsedAt    *time.Time     `json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	User      User           `gorm:"foreignKey:UserID"`
}
