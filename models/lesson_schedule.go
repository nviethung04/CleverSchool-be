package models

import (
	"time"
)

type LessonSchedule struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	CourseID      int64     `json:"course_id"`
	LessonID      int64     `json:"lesson_id"`
	LessonPlanID  int64     `json:"lesson_plan_id"`
	ShiftID       *int64    `json:"shift_id"` // nullable
	ScheduledDate time.Time `json:"scheduled_date"`
	WeekID        int64     `json:"week_id"`
	SortPosition  int       `json:"sort_position"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Course     *Course     `gorm:"foreignKey:CourseID"`
	Week       *Week       `gorm:"foreignKey:WeekID"`
	StudyShift *StudyShift `gorm:"foreignKey:ShiftID"`
	Lesson     Lesson      `gorm:"foreignKey:LessonID"`
}
