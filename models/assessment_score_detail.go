package models

import (
	"time"

	"gorm.io/gorm"
)

type AssessmentScoreDetail struct {
	ID                int64          `gorm:"primaryKey" json:"id"`
	AssessmentScoreID int64          `gorm:"column:assessment_score_id" json:"assessment_score_id"`
	CriterionID       int64          `gorm:"column:criterion_id" json:"criterion_id"`
	SubcriterionID    *int64         `gorm:"column:subcriterion_id" json:"subcriterion_id"`
	Score             float64        `gorm:"column:score;type:numeric(8,2)" json:"score"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	CreatedBy         int64          `gorm:"column:created_by" json:"created_by"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy         int64          `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy         int64          `gorm:"column:deleted_by" json:"deleted_by"`
}

func (AssessmentScoreDetail) TableName() string {
	return "assessment_score_details"
}
