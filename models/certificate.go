package models

import (
	"time"

	"gorm.io/gorm"
)

type Certificate struct {
	ID              int64     `json:"id"`
	CourseId       int64      `json:"course_id" gorm:"index"`
	UserId          int64     `json:"user_id"`
	Name            string    `json:"name"`
	IssuedBy        string    `json:"issued_by"`
	IssuedDate      time.Time `json:"issued_date"`
	ExpiryDate      time.Time `json:"expiry_date"`
	CertificateCode string    `json:"certificate_code"`
	Description     string    `json:"description"`
	FileInfo        MediaInfo `gorm:"type:jsonb" json:"file_info"`
	Status          bool      `gorm:"not null" json:"status"`
	Grade           string         `json:"grade"`
	Rating          float32        `json:"rating"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `json:"deleted_by"`
}
