package dto

import (
	"be-cleverschool/models"
	"time"

	"gorm.io/gorm"
)

type HomeworkDTO struct {
	ID                      int64            `gorm:"primaryKey" json:"id"`
	LessonID                int64            `json:"lesson_id"`
	Name                    string           `json:"name"`
	Status                  int16            `json:"status"`
	Description             string           `json:"description"`
	MaxScore                float64          `json:"max_score"`
	IsAssigned              *bool            `json:"is_assigned"`
	IsRandomQuestion        bool            `json:"is_random_question"`
	CoverImage              string           `json:"cover_image"`
	CoverImageInfo          models.MediaInfo `gorm:"type:jsonb" json:"cover_image_info"`
	TotalQuestion           int32            `json:"total_question"`
	QuestionCompleted       int32            `json:"question_completed"`
	LastQuestionIDCompleted int64            `json:"last_question_id_completed"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
	CreatedBy               int64            `json:"created_by"`
	UpdatedBy               int64            `json:"updated_by"`
	DeletedAt               gorm.DeletedAt   `gorm:"column:deleted_at"`
	DeletedBy               int64            `gorm:"column:deleted_by"`
	AssignedAt              time.Time        `json:"assigned_at"`
	AssignedBy              int64            `json:"assigned_by"`
	ProgramID               int64            `json:"program_id"`
	ObjectTitle             string           `gorm:"size:255;not null" json:"object_title"`
	CloneInfo               *models.CloneInfo `gorm:"type:jsonb" json:"clone_info"`
	QuestionForm            string           `gorm:"type:question_form_enum;not null"`
	FileInfos               models.MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
}

