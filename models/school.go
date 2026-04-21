package models

import (
	"time"

	"gorm.io/gorm"
)

type School struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	ShortName      string    `json:"short_name"`
	Type           string    `json:"type"`
	WardCode       string    `json:"ward_code"`
	AddressVN      string    `json:"address_vn"`
	AddressEN      string    `json:"address_en"`
	ParentSchoolID *int64    `json:"parent_school_id,omitempty"` // nullable FK
	ContactName    string    `json:"contact_name"`
	ContactPhone   string    `json:"contact_phone"`
	ContactMail    string    `json:"contact_mail"`
	Status         bool      `gorm:"not null" json:"status"`
	LogoInfo       MediaInfo `gorm:"type:jsonb" json:"logo_info"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `json:"deleted_by"`

	StudentCount int32 `gorm:"null" json:"student_count"`
	ClassCount   int32 `gorm:"null" json:"class_count"`

	WardName     string `json:"ward_name" gorm:"-"`
	ProvinceName string `json:"province_name" gorm:"-"`
	ProvinceCode string `json:"province_code" gorm:"-"`

	Ward Ward `gorm:"foreignKey:WardCode;references:Code" json:"ward"`
}
