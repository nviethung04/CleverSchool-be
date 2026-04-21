package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	QuestionFormQuestionType   = "question"
	QuestionFormWriteType     = "write"
)

type Homework struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	ProgramId       int64     `gorm:"null" json:"program_id"`
	Name           string         `json:"name"`
	ObjectTitle            string    `gorm:"size:255;not null" json:"object_title"`
	Status         int16          `json:"status"`
	Description    string         `json:"description"`
	MaxScore       float64        `json:"max_score"`
	CoverImageInfo MediaInfo      `gorm:"type:jsonb" json:"cover_image_info"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	CreatedBy      int64          `json:"created_by"`
	UpdatedBy      int64          `json:"updated_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
	DeletedBy      int64          `gorm:"column:deleted_by"`
	TotalQuestions int32 `gorm:"column:total_questions" json:"total_questions"`
	IsRandomQuestion bool           `gorm:"column:is_random_question" json:"is_random_question"`

	CloneInfo *CloneInfo `gorm:"type:jsonb" json:"clone_info"`
    Lessons   []Lesson   `gorm:"many2many:homework_ref_lessons"`
	HomeworkRefLessons []HomeworkRefLesson `gorm:"foreignKey:HomeworkId"`

	QuestionForm string `gorm:"type:question_form_enum;not null"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
}
