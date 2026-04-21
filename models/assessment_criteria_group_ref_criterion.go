package models

import (
	"time"

	"gorm.io/gorm"
)

type AssessmentCriteriaGroupRefCriterion struct {
	ID                        int64               `gorm:"primaryKey" json:"id"`
	AssessmentCriteriaGroupID int64               `gorm:"column:assessment_criteria_group_id" json:"assessment_criteria_group_id"`
	AssessmentCriteriaID      int64               `gorm:"column:assessment_criteria_id" json:"assessment_criteria_id"`
	CreatedAt                 time.Time           `gorm:"column:created_at" json:"created_at"`
	CreatedBy                 int64               `gorm:"column:created_by" json:"created_by"`
	UpdatedAt                 time.Time           `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy                 int64               `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt                 gorm.DeletedAt      `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy                 int64               `gorm:"column:deleted_by" json:"deleted_by"`
	AssessmentCriterion       AssessmentCriterion `gorm:"foreignKey:AssessmentCriteriaID"`
}

func (AssessmentCriteriaGroupRefCriterion) TableName() string {
	return "assessment_criteria_group_ref_criteria"
}
