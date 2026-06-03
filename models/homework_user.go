package models

import (
	"time"
)

type HomeworkUser struct {
	ID                      int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID              int64     `gorm:"column:homework_id" json:"homework_id"`
	LessonID                int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID                  int64     `gorm:"column:user_id" json:"user_id"`
	LastQuestionIDCompleted int64     `gorm:"column:last_question_id_completed" json:"last_question_id_completed"`
	QuestionsCompleted      int64     `gorm:"column:questions_completed" json:"questions_completed"`
    Score                   float64   `gorm:"column:score" json:"score"`
    Ratio                   float64   `gorm:"column:ratio" json:"ratio"`
    HasManualScoring        bool      `gorm:"column:has_manual_scoring" json:"has_manual_scoring"`
    StatusScoring           int16     `gorm:"column:status_scoring" json:"status_scoring"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
	CreatedAt               time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (HomeworkUser) TableName() string {
	return "homework_users"
}
