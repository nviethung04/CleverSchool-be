package models

import "time"

type UserCourse struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId      int64     `json:"user_id"`
	CourseId    int64     `json:"course_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsCurrent   bool      `gorm:"column:is_current" json:"is_current"`
	Course      *Course   `gorm:"foreignKey:CourseId"`
	MainTeacher bool      `gorm:"column:main_teacher" json:"main_teacher"`
}
