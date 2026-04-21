package models

import (
	"time"

	"gorm.io/gorm"
)

type TrainingLevel struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	SortPosition int32  `json:"sort_position"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `json:"deleted_by"`
}
