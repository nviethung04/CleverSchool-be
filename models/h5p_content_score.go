package models

import (
	"time"
)

type H5pContentScore struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentId string    `gorm:"column:content_id"`
	Score     float32   `gorm:"column:score"`
	MaxScore  float32   `gorm:"column:max_score"`
	Opened    time.Time `gorm:"column:opened"`
	Finished  time.Time `gorm:"column:finished"`
	Time *time.Time `gorm:"column:time" json:"time"`
	UserId    int64    `gorm:"column:user_id"`
	CreatedAt time.Time `json:"created_at"`
}
