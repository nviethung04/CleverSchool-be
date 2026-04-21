package models

type LessonRefTopic struct {
	LessonID int64 `gorm:"primaryKey"`
	TopicID  int64 `gorm:"primaryKey"`
}
