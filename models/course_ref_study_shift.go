package models

import (
	"time"
)

type CourseRefStudyShift struct {
	ID        int64 `json:"id" gorm:"primaryKey"`
	CourseID  int64 `json:"course_id"`
	ShiftID   int64 `json:"shift_id"`
	DayOfWeek int32 `json:"day_of_week"`

	CreatedAt time.Time `json:"created_at"`

	StudyShift StudyShift `gorm:"foreignKey:ShiftID"`
}
