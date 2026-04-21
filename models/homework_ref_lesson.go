package models

import "time"

type HomeworkRefLesson struct {
	ID         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	HomeworkId   int64 `gorm:"not null" json:"homework_id"`
	LessonId int64 `gorm:"not null" json:"lesson_id"`
    AssignedAt     *time.Time     `json:"assigned_at"`
    AssignedBy     *int64         `json:"assigned_by"`
	CourseId       int64     `gorm:"null" json:"course_id"`
	Course  Course   `gorm:"foreignKey:CourseId"`
}
