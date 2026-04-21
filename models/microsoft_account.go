package models

import (
	"time"

	"gorm.io/gorm"
)

// MicrosoftAccount represents user's connected Microsoft account
type MicrosoftAccount struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`
	MicrosoftUserID string         `json:"microsoft_user_id" gorm:"not null"`
	AccessToken     string         `json:"access_token" gorm:"not null"`
	RefreshToken    string         `json:"refresh_token" gorm:"not null"`
	ExpiresAt       time.Time      `json:"expires_at" gorm:"not null"`
	AccountEmail    string         `json:"account_email" gorm:"not null"`
	TenantID        string         `json:"tenant_id"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"DeletedAt" gorm:"index"`

	// Relationships
	User User `json:"User" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName specifies the table name for MicrosoftAccount model
func (MicrosoftAccount) TableName() string {
	return "microsoft_accounts"
}
