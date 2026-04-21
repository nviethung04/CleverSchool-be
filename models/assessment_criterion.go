package models

import (
	"time"

	"gorm.io/gorm"
)

type AssessmentCriterion struct {
	ID                    int64                    `gorm:"primaryKey" json:"id"`
	SubjectID             int64                    `gorm:"column:subject_id" json:"subject_id"`
	Name                  string                   `json:"name"`
	Description           string                   `json:"description"`
	HasSubcriteria        bool                     `gorm:"column:has_subcriteria" json:"has_subcriteria"`
	MaxScore              float32                  `gorm:"column:max_score" json:"max_score"`
	CreatedAt             time.Time                `json:"created_at"`
	UpdatedAt             time.Time                `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy             int64                    `json:"created_by"`
	UpdatedBy             int64                    `json:"updated_by"`
	DeletedAt             gorm.DeletedAt           `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy             int64                    `gorm:"column:deleted_by" json:"deleted_by"`
	AssessmentSubcriteria []AssessmentSubcriterion `gorm:"foreignKey:CriterionID"`
}

func (AssessmentCriterion) TableName() string {
	return "assessment_criteria"
}
