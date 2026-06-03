package models

import (
	"time"

	"gorm.io/gorm"
)

type Program struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectId       int64     `gorm:"null" json:"subject_id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	Description     string    `gorm:"type:text;not null" json:"description"`
	ImageInfo       MediaInfo `gorm:"type:jsonb" json:"image_info"`
	Status          bool      `gorm:"null" json:"status"`
	Target          string    `gorm:"null" json:"target"`
	Duration        int       `json:"duration"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64          `gorm:"null" json:"created_by"`
	UpdatedBy int64          `gorm:"null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	Chapters    []Chapter    `gorm:"foreignKey:ProgramId" json:"chapters"`
    Courses     []Course     `gorm:"foreignKey:ProgramId" json:"courses"`
	Subject     Subject      `gorm:"foreignKey:SubjectId"`
}
