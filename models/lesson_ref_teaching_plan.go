package models

type LessonRefTeachingPlan struct {
	LessonId       int64 `gorm:"not null" json:"lesson_id"`
	TeachingPlanId int64 `gorm:"not null" json:"teaching_plan_id"`
}
