package models

import "time"

type LessonCompletion struct {
	StudentID   int64     `gorm:"not null"`
	LessonID    int64     `gorm:"not null"`
	CompletedAt time.Time `json:"completed_at"`
}
