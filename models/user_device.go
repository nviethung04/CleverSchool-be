package models

import "time"

// UserDevice represents a user's device with FCM token
type UserDevice struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	UserID      int64      `gorm:"not null;index" json:"user_id"`
	DeviceToken string     `gorm:"type:varchar(500);not null;uniqueIndex" json:"device_token"`
	DeviceType  string     `gorm:"type:varchar(20);not null;index" json:"device_type"`
	DeviceName  *string    `gorm:"type:varchar(100)" json:"device_name"`
	Platform    *string    `gorm:"type:varchar(20)" json:"platform"`
	AppVersion  *string    `gorm:"type:varchar(20)" json:"app_version"`
	OsVersion   *string    `gorm:"type:varchar(20)" json:"os_version"`
	IsActive    bool       `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastUsedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"last_used_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (UserDevice) TableName() string {
	return "user_devices"
}

// Constants for device types
const (
	DeviceTypeIOS     = "ios"
	DeviceTypeAndroid = "android"
	DeviceTypeWeb     = "web"
)
