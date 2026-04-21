package models

import "time"

type MeetingNotification struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	MeetingID   uint       `json:"meeting_id" gorm:"index;not null"`
	UserID      uint       `json:"user_id" gorm:"index;not null"`
	CourseID    *uint      `json:"course_id" gorm:"index"`
	LessonID    *uint      `json:"lesson_id" gorm:"index"`
	Provider    string     `json:"provider" gorm:"size:20;default:'microsoft'"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReadAt      *time.Time `json:"read_at"`
}

func (MeetingNotification) TableName() string {
	return "meeting_notifications"
}
