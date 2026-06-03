package models

import "time"

type UserCourse struct {
	UserId      int64     `json:"user_id"`
	CourseId    int64     `json:"course_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsCurrent   bool      `json:"is_current"`
	Course      *Course   `gorm:"foreignKey:CourseId"`
	MainTeacher bool      `json:"main_teacher"`
}
