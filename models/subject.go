package models

import "time"

type Subject struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text;not null" json:"description"`
	Status      bool   `gorm:"not null" json:"status"`

	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64      `gorm:"null" json:"created_by"`
	UpdatedBy int64      `gorm:"null" json:"updated_by"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	DeletedBy int64      `gorm:"null" json:"deleted_by,omitempty"`
}
