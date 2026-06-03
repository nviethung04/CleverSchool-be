package models

import (
	"time"

	"gorm.io/gorm"
)

type SourceQuestion struct {
	ID    int64  `gorm:"primaryKey"`
	Skill string `gorm:"type:SKILL_ENUM;not null"`
	Level string `gorm:"type:LEVEL_ENUM;not null"`

	Title   string
	Content string
	Status  bool `gorm:"not null" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
