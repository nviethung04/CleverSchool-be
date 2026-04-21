package models

import (
	"time"
)

type Notice struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	Cover       *string    `json:"cover"`
	Title       string     `json:"title" gorm:"type:varchar(500);not null"`
	Description *string    `json:"description"`
	Content     *string    `json:"content"`
	Status      bool       `json:"status" gorm:"default:false;not null"`
	Type        string     `json:"type" gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	CreatedBy   *int64     `json:"created_by"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	UpdatedBy   *int64     `json:"updated_by"`
	DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`

	NotificationLogs []NotificationLog `gorm:"foreignKey:NoticeID" json:"logs"`
}

func (Notice) TableName() string {
	return "notices"
}

const (
	NoticeTypeSystem  = "system"
	NoticeTypeSchool  = "school"
	NoticeTypeClass   = "class"
	NoticeTypeCourse  = "course"
	NoticeTypeStudent = "student"
	NoticeTypeUser    = "user"
)

const (
	NoticeStatusDraft     = false
	NoticeStatusPublished = true
)
