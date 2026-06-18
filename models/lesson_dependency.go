package models

type LessonDependency struct {
	LessonID     int64 `gorm:"primaryKey"`
	DependencyID int64 `gorm:"column:dependency_lesson_id;primaryKey"`
}
