package models

type LessonRefTag struct {
	LessonID int64 `gorm:"primaryKey"`
	TagID    int64 `gorm:"primaryKey"`
}
