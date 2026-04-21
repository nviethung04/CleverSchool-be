package models

import (
	"time"

	"gorm.io/gorm"
)

type ZoomAccount struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64          `gorm:"not null;index" json:"user_id"`
	ZoomUserID   string         `gorm:"size:255" json:"zoom_user_id"`
	AccessToken  string         `gorm:"type:text" json:"access_token"`
	RefreshToken string         `gorm:"type:text" json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at"`
	AccountEmail string         `gorm:"size:255" json:"account_email"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations
	User User `gorm:"foreignKey:UserID"`
}

func (ZoomAccount) TableName() string {
	return "zoom_accounts"
}
