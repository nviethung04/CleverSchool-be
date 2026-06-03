package models

import (
	"time"

	"gorm.io/gorm"
)

type LessonPlanPart struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	LessonPlanID   int64          `json:"lesson_plan_id"`
	ProgramId       int64     `gorm:"null" json:"program_id"`
	CourseID       int64     `gorm:"null" json:"course_id"`
	Title          string         `json:"title"`
	ObjectTitle            string    `gorm:"size:255;not null" json:"object_title"`
	Tag            string         `json:"tag"`
	CoverImageInfo MediaInfo      `gorm:"type:jsonb" json:"cover_image_info"`
	SortPosition   int16          `json:"sort_position"`
	Time           int64          `json:"time"`
	IsClasswork    bool           `json:"is_classwork"`
	FileType       string         `json:"file_type"`
	LinkInfo       MediaInfo      `gorm:"type:jsonb" json:"link_info"`
	LinkType       string         `json:"link_type"`
	GuideTeacher   string         `json:"guide_teacher"`
	GuideStudent   string         `json:"guide_student"`
	File           string         `json:"file"`
	MaxScore       float64        `json:"max_score"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	CreatedBy      int64          `json:"created_by"`
	UpdatedBy      int64          `json:"updated_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy      int64          `gorm:"column:deleted_by"`

	CloneInfo *CloneInfo `gorm:"type:jsonb" json:"clone_info"`
}
