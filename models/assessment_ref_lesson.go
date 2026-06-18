package models

type AssessmentRefLesson struct {
	ID           int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	AssessmentId int64 `gorm:"not null" json:"assessment_id"`
	LessonId     int64 `gorm:"not null" json:"lesson_id"`
	CourseId     int64 `gorm:"null" json:"course_id"`
}
