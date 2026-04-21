package models

import (
	"time"

	"gorm.io/datatypes"
)

type DeviceInfo struct {
	Desktop int64 `json:"desktop"`
	Tablet  int64 `json:"tablet"`
	Mobile  int64 `json:"mobile"`
	All     int64 `json:"all"`
}

type UsageTime struct {
	Student int64 `json:"student"`
	Teacher int64 `json:"teacher"`
	Admin   int64 `json:"admin"`
	All     int64 `json:"all"`
}

type AverageUsed struct {
	MinutesPerSession float64 `json:"minutes_per_session"`
	DailyActiveUsers float64 `json:"daily_active_users"`
	EngagementRate   float64 `json:"engagement_rate"`
	ReturningRate     float64 `json:"returning_rate"`
}

type ActivityIds struct {
	Ids []int64 `json:"ids"`
}

type HistoryUse struct {
	ID        int64          `gorm:"primaryKey"`
	Type      string         `gorm:"type:varchar(10);not null"` // 'week' hoặc 'month'
	Date      time.Time      `gorm:"type:date;not null;index"`
	UsageTime datatypes.JSON `gorm:"type:jsonb;not null"`
	Device    datatypes.JSON `gorm:"type:jsonb;not null"`
	AverageUsed datatypes.JSON `gorm:"type:jsonb;not null"`
	ActivityIds datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
