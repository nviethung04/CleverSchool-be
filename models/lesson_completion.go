package models

import "time"

type LessonCompletion struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	LessonID    int64     `gorm:"not null;index"`
	StudentID   int64     `gorm:"not null;column:user_id;index"`
	CompletedAt time.Time `json:"completed_at"`
}

func (LessonCompletion) TableName() string {
	return "lesson_completions"
}
