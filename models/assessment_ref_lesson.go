package models

import "time"

type AssessmentRefLesson struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AssessmentId int64      `gorm:"not null" json:"assessment_id"`
	LessonId     int64      `gorm:"not null" json:"lesson_id"`
	CourseId     int64      `gorm:"column:course_id" json:"course_id"`
	AssignedAt   *time.Time `json:"assigned_at"`
	AssignedBy   *int64     `json:"assigned_by"`
	Course       Course     `gorm:"foreignKey:CourseId"`
}
