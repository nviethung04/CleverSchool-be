package models

import (
	"time"

	"gorm.io/gorm"
)

type LessonPlan struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	ProgramId      int64     `gorm:"null" json:"program_id"`
	AuthorId       int64     `gorm:"null" json:"author_id"`
	Name           string    `json:"name"`
	ObjectTitle    string    `gorm:"size:255;not null" json:"object_title"`
	Description    string    `json:"description"`
	CoverImageInfo MediaInfo `gorm:"type:jsonb" json:"cover_image_info"`
	Status         int       `json:"status"`
	SortPosition   int       `json:"sort_position"`
	TotalTime      int       `json:"total_time"`
	Views          int       `json:"views"`

	Complete *LessonPlanComplete `gorm:"foreignKey:LessonPlanID"`
	Author   *User               `gorm:"foreignKey:AuthorId"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	CloneInfo *CloneInfo `gorm:"type:jsonb" json:"clone_info"`
	Lessons   []Lesson   `gorm:"many2many:lesson_plan_ref_lessons"`
	// LessonPlanRefLessons []LessonPlanRefLesson `gorm:"foreignKey:LessonPlanId"`
}
