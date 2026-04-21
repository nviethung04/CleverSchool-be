package models

import "time"

type SessionStatus string

const (
	SessionPlanned   SessionStatus = "planned"
	SessionOngoing   SessionStatus = "ongoing"
	SessionCompleted SessionStatus = "completed"
)

type Session struct {
	ID              int64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string        `gorm:"size:255;not null" json:"name"`
	StartAt         time.Time     `gorm:"not null" json:"start_at"`
	Note            string        `gorm:"size:255" json:"note"`
	Status          SessionStatus `gorm:"type:enum('planned','ongoing','completed');default:'planned'" json:"status"`
	AttendanceCount int           `gorm:"default:0" json:"attendance_count"`
}
