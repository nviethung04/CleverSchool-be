package models

import (
	"time"

	"gorm.io/gorm"
)

type Semester struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	BeginDate   time.Time `gorm:"not null" json:"begin_date"`
	Status      bool      `gorm:"default:true" json:"status"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	PreviousSemesterId   int64       `gorm:"default:0" json:"previous_semester_id"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64          `gorm:"null" json:"created_by"`
	UpdatedBy int64          `gorm:"null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	StartWeekId int64 `gorm:"null" json:"start_week_id"`
	EndWeekId   int64 `gorm:"null" json:"end_week_id"`

	// Relationships
	Courses  []Course  `gorm:"many2many:course_ref_semesters" json:"courses"`
	Holidays []Holiday     `gorm:"many2many:semester_ref_holidays"`
}
