package models

import "time"

type AssessmentRefCriterion struct {
	ID                    int64               `gorm:"primaryKey" json:"id"`
	AssessmentID          int64               `gorm:"column:assessment_id" json:"assessment_id"`
	AssessmentCriterionID int64               `gorm:"column:assessment_criterion_id" json:"assessment_criterion_id"`
	CreatedAt             time.Time           `gorm:"column:created_at" json:"created_at"`
	CreatedBy             int64               `gorm:"column:created_by" json:"created_by"`
	UpdatedAt             time.Time           `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy             int64               `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt             *time.Time          `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy             int64               `gorm:"column:deleted_by" json:"deleted_by"`
	AssessmentCriterion   AssessmentCriterion `gorm:"foreignKey:AssessmentCriterionID"`
}

func (AssessmentRefCriterion) TableName() string {
	return "assessment_ref_criteria"
}
