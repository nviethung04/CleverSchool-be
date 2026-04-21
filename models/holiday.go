package models

import (
	"time"

	"gorm.io/gorm"
)

type Holiday struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`
	Type        string    `gorm:"size:100;not null;default:'holiday'" json:"type"` // holiday, exam_break, etc.
	Status      bool      `gorm:"default:true" json:"status"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64          `gorm:"null" json:"created_by"`
	UpdatedBy int64          `gorm:"null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
	WeekId    int64          `gorm:"null" json:"week_id"`

	// Relationships
	Semesters []Semester     `gorm:"many2many:semester_ref_holidays"`
}
