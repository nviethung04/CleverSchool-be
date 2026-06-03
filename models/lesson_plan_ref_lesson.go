package models

type LessonPlanRefLesson struct {
	LessonPlanId   int64 `gorm:"not null" json:"lesson_plan_id"`
	LessonId int64 `gorm:"not null" json:"lesson_id"`
	CourseId       int64     `gorm:"null" json:"course_id"`
	Course  Course   `gorm:"foreignKey:CourseId"`
}
