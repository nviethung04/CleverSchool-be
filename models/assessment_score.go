package models

import (
	"time"

	"gorm.io/gorm"
)

type AssessmentScore struct {
	ID           int64          `gorm:"primaryKey" json:"id"`
	AssessmentID int64          `gorm:"column:assessment_id" json:"assessment_id"`
	StudentID    int64          `gorm:"column:student_id" json:"student_id"`
	LessonID     int64          `gorm:"column:lesson_id" json:"lesson_id"`
	CourseID     int64          `gorm:"column:course_id" json:"course_id"`
	TotalScore   float64        `gorm:"column:total_score;type:numeric(8,2)" json:"total_score"`
	StatusScored int            `gorm:"column:status_scored" json:"status_scored"`
	IsLate       bool           `gorm:"column:is_late" json:"is_late"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	CreatedBy    int64          `gorm:"column:created_by" json:"created_by"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	UpdatedBy    int64          `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	DeletedBy    int64          `gorm:"column:deleted_by" json:"deleted_by"`
	FileInfos    MediaInfos     `gorm:"column:file_infos" json:"file_infos"`
}

func (AssessmentScore) TableName() string {
	return "assessment_scores"
}
