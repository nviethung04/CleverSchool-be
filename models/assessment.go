package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type AssessmentCriteriaIDs []int64

func (ids *AssessmentCriteriaIDs) Scan(value interface{}) error {
	if value == nil {
		*ids = []int64{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan AssessmentCriteriaIDs: wrong type %T", value)
	}
	if len(bytes) == 0 {
		*ids = []int64{}
		return nil
	}
	return json.Unmarshal(bytes, ids)
}

func (ids AssessmentCriteriaIDs) Value() (driver.Value, error) {
	if ids == nil {
		return json.Marshal([]int64{})
	}
	return json.Marshal(ids)
}

type Assessment struct {
	ID                         int64                 `gorm:"primaryKey" json:"id"`
	Name                       string                `json:"name"`
	Description                string                `json:"description"`
	Type                       string                `gorm:"column:type" json:"type"`
	ProgramId                  int64                 `gorm:"null" json:"program_id"`
	SubjectId                  int64                 `gorm:"null" json:"subject_id"`
	AssessmentCriteriaGroupId  int64                 `gorm:"null" json:"assessment_criteria_group_id"`
	StudyReportCriteriaId      int64                 `gorm:"null" json:"study_report_criteria_id"`
	AssessmentCriteriaIds      AssessmentCriteriaIDs `gorm:"type:jsonb;default:'[]'" json:"assessment_criteria_ids"`
	FileInfos                  MediaInfos            `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
	Subject                    *Subject              `gorm:"foreignKey:SubjectId"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
