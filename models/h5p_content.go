package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type H5pContent struct {
	ID           int64          `json:"id" gorm:"primaryKey"`
	ContentID        string         `gorm:"type:text" json:"content_id"`
	Title        string         `gorm:"type:text" json:"title"`
	Library        string         `gorm:"type:text" json:"library"`
	Parameters   datatypes.JSON `gorm:"type:jsonb" json:"parameters"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	SortPosition int            `json:"sort_position"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
