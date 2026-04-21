package models

type CourseSchool struct {
	CourseId int64   `json:"course_id"`
	SchoolId int64   `json:"school_id"`
	Course   *Course `gorm:"foreignKey:CourseId"`
}
