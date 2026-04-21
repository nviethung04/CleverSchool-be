package models

import (
	"time"

	"gorm.io/gorm"
)

type Assessment struct {
	ID                    int64          `gorm:"primaryKey" json:"id"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	Type                  string         `gorm:"type:assessment_type_enum" json:"type"`
	ProgramID             int64          `gorm:"column:program_id" json:"program_id"`
	SubjectID             int64          `gorm:"column:subject_id" json:"subject_id"`
	StudyReportCriteriaID int64          `gorm:"column:study_report_criteria_id" json:"study_report_criteria_id"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `gorm:"column:update_at" json:"updated_at"`
	CreatedBy             int64          `json:"created_by"`
	UpdatedBy             int64          `json:"updated_by"`
	DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy             int64          `gorm:"column:deleted_by" json:"deleted_by"`
	FileInfos             MediaInfos     `gorm:"column:file_infos" json:"file_infos"`
	HasFile               bool           `gorm:"column:has_file" json:"has_file"`

	AssessmentRefLessons  []AssessmentRefLesson    `gorm:"foreignKey:AssessmentId"`
	AssessmentRefCriteria []AssessmentRefCriterion `gorm:"foreignKey:AssessmentID"`

	// Các trường được populate từ join, không lưu trong DB
	AssessmentCriteriaGroupID      int64  `gorm:"->" json:"assessment_criteria_group_id"`
	AssessmentCriteriaGroupName    string `gorm:"->" json:"assessment_criteria_group_name"`
	AssessmentCriteriaGroupHasFile bool   `gorm:"->" json:"assessment_criteria_group_has_file"`
	StudyReportCriteriaName        string `gorm:"->" json:"study_report_criteria_name"`
	SubjectName                    string `gorm:"->" json:"subject_name"`
	// Các trường từ assessment_ref_lessons (khi có course_id)
	LessonID    int64  `gorm:"->" json:"lesson_id"`
	LessonTitle string `gorm:"->" json:"lesson_title"`
	IsAssigned  bool   `gorm:"->" json:"is_assigned"`

	PublishCourseIds []int64 `gorm:"-" json:"publish_course_ids"`
}
