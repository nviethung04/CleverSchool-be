package models

import "time"

type UserClass struct {
	UserId    int64     `json:"user_id"`
	ClassId   int64     `json:"class_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsCurrent bool      `json:"is_current"`
	Class     *Class    `gorm:"foreignKey:ClassId"`
}
