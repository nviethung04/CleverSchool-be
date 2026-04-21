package models

import (
	"time"

	"gorm.io/gorm"
)

// Feedback status constants
const (
	FeedbackStatusPending   int64 = 1
	FeedbackStatusApproved  int64 = 2
	FeedbackStatusRejected  int64 = 3
	FeedbackStatusResolved  int64 = 4
)

// Feedback type constants
const (
	FeedbackTypeGeneral    int64 = 1
	FeedbackTypeBug        int64 = 2
	FeedbackTypeFeature    int64 = 3
	FeedbackTypeOther      int64 = 4
)

type Feedback struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	UserID    *int64         `json:"user_id"`
	Content   string         `json:"content"`
	Title     string         `json:"title"`
	FileInfo  MediaInfo      `gorm:"type:jsonb" json:"file_info"`
	RoleID    *int64         `json:"role_id"`
	Status    *int64         `json:"status"`
	Response  string         `json:"response"`
	Note      string         `json:"note"`
	Type      *int64         `json:"type"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	User User `gorm:"foreignKey:UserID"`
	Role Role `gorm:"foreignKey:RoleID"`
} 