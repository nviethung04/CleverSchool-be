package models

type HeadingRefLesson struct {
	ID        int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	HeadingId int64 `gorm:"not null" json:"heading_id"`
	LessonId  int64 `gorm:"not null" json:"lesson_id"`
}

func (HeadingRefLesson) TableName() string {
	return "heading_ref_lessons"
}
