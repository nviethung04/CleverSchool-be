package models

import (
	"time"

	"gorm.io/gorm"
)

type Grade struct {
	ID     int64  `gorm:"primaryKey" json:"id"`
	Number int32  `json:"number"`
	NameVN string `json:"name_vn"`
	NameEN string `json:"name_en"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
