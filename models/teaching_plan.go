package models

import (
	"time"

	"gorm.io/gorm"
)

type TeachingPlan struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	FileInfo    MediaInfo `gorm:"type:jsonb" json:"file_info"`
	FileType    string    `json:"file_type"`

	Status       bool      `json:"status"`
	ApprovedNote string    `json:"approved_note"`
	ApprovedBy   int64     `json:"approved_by"`
	ApprovedAt   time.Time `json:"approved_at"`

	Lessons []Lesson `gorm:"many2many:lesson_ref_teaching_plans;joinForeignKey:teaching_plan_id;joinReferences:lesson_id"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `json:"deleted_by"`
}
