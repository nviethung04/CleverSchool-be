package models

import (
	"time"
)

type HomeworkUserSkipQuestion struct {
	ID            int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID    int64     `gorm:"column:homework_id" json:"homework_id"`
	LessonID      int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID        int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID    int64     `gorm:"column:question_id" json:"question_id"`
	DidItAgain    bool      `gorm:"column:did_it_again" json:"did_it_again"`
	DidItAgainAt  *time.Time `gorm:"column:did_it_again_at" json:"did_it_again_at"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (HomeworkUserSkipQuestion) TableName() string {
	return "homework_user_skip_questions"
}
