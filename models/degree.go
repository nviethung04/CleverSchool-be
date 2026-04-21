package models

import (
	"time"

	"gorm.io/gorm"
)

type Degree struct {
	ID             int64     `json:"id"`
	UserId         int64     `json:"user_id"`
	Name           string    `json:"name"`
	Major          string    `json:"major"`
	Institution    string    `json:"institution"`
	GraduationYear int       `json:"graduation_year"`
	DegreeLevel    string    `json:"degree_level"`
	DegreeCode     string    `json:"degree_code"`
	ReceivedDate   time.Time `json:"received_date"`
	Note           string    `json:"note"`
	Status         bool      `gorm:"not null" json:"status"`
	FileInfo       MediaInfo `gorm:"type:jsonb" json:"file_info"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `json:"deleted_by"`
}
