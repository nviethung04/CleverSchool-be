package models

import (
	"time"

	"gorm.io/gorm"
)

type EmployeePosition struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description"`
	Status      bool   `gorm:"not null" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `json:"deleted_by"`
}
