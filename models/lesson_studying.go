package models

import "time"

type LessonStudying struct {
	StudentID  int64     `gorm:"not null"`
	LessonID   int64     `gorm:"not null"`
	StudyingAt time.Time `json:"studying_at"`
}

func (LessonStudying) TableName() string {
	return "lesson_studying"
}
