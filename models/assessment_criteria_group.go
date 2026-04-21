package models

import (
	"time"

	"gorm.io/gorm"
)

type AssessmentCriteriaGroup struct {
	ID          int64                                 `gorm:"primaryKey" json:"id"`
	Name        string                                `json:"name"`
	SubjectID   int64                                 `gorm:"column:subject_id" json:"subject_id"`
	HasFile     bool                                  `gorm:"column:has_file" json:"has_file"`
	CreatedAt   time.Time                             `json:"created_at"`
	UpdatedAt   time.Time                             `json:"updated_at"`
	CreatedBy   int64                                 `json:"created_by"`
	UpdatedBy   int64                                 `json:"updated_by"`
	DeletedAt   gorm.DeletedAt                        `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy   int64                                 `gorm:"column:deleted_by" json:"deleted_by"`
	RefCriteria []AssessmentCriteriaGroupRefCriterion `gorm:"foreignKey:AssessmentCriteriaGroupID"`
}

func (AssessmentCriteriaGroup) TableName() string {
	return "assessment_criteria_groups"
}
