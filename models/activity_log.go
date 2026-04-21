package models

import (
	"time"

	"gorm.io/datatypes"
)

type ActivityLog struct {
	ID           uint   `gorm:"primaryKey"`
	UserID       *int   `gorm:"index"`
	RoleID       *int   `gorm:"index"`
	Method       string `gorm:"type:varchar(10);not null"`
	Path         string `gorm:"type:text;not null"`
	StatusCode   int    `gorm:"not null"`
	ClientIP     string `gorm:"type:varchar(45)"`
	Device     string `gorm:"type:varchar(20)"`
	LatencyMs    int
	QueryParams  datatypes.JSON `gorm:"type:jsonb"`
	RequestBody  datatypes.JSON `gorm:"type:jsonb"`
	ResponseBody datatypes.JSON `gorm:"type:jsonb"`
	Attribute    datatypes.JSON `gorm:"type:jsonb"`
	SessionID    *string        `gorm:"type:text"`
	Agent        *string        `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
}
