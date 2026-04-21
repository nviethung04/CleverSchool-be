package models

import (
	"time"
)

type LessonPlanComplete struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `json:"user_id"`
	LessonPlanID int64     `json:"lesson_plan_id"`
	CompletedAt  time.Time `json:"completed_at"`
}
