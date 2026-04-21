package models

import (
	"time"

	"gorm.io/gorm"
)

// GoogleAccount represents user's connected Google account
type GoogleAccount struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	GoogleUserID string         `json:"google_user_id" gorm:"not null"`
	AccessToken  string         `json:"access_token" gorm:"not null"`
	RefreshToken string         `json:"refresh_token" gorm:"not null"`
	ExpiresAt    time.Time      `json:"expires_at" gorm:"not null"`
	AccountEmail string         `json:"account_email" gorm:"not null"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"DeletedAt" gorm:"index"`

	// Relationships
	User User `json:"User" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName specifies the table name for GoogleAccount model
func (GoogleAccount) TableName() string {
	return "google_accounts"
}
