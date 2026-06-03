package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	ClonedQuestionTypeHomework       = "homework"
	ClonedQuestionTypeExam           = "exam"
	ClonedQuestionTypeExercise       = "exercise"
	ClonedQuestionTypeLessonPlanPart = "lesson_plan_part"
	ClonedQuestionTypeLevelTest      = "level_test"
	ClonedQuestionTypeContestRound   = "contest_round"
)

type ClonedQuestion struct {
	ID             int64 `gorm:"primaryKey"`
	ProgramId      int64 `gorm:"null" json:"program_id"`
	AssignmentID   int64
	AssignmentType string
	Questions      datatypes.JSON

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	CloneInfo *CloneInfo `gorm:"type:jsonb" json:"clone_info"`
}
