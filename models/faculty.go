package models

import "time"

func (Faculty) TableName() string {
	return "faculties"
}

type Faculty struct {
	ID          int64   `gorm:"primaryKey;autoIncrement" json:"id"`
	SchoolId    int64   `gorm:"null" json:"school_id"`
	Name        string  `gorm:"size:255;not null" json:"name"`
	Code        string  `gorm:"size:50;unique" json:"code"`
	Description string  `json:"description"`
	Status      bool    `gorm:"null" json:"status"`
	School      *School `gorm:"foreignKey:SchoolId"`

	NameHead   string    `gorm:"size:255" json:"name_head"`
	AvatarInfo MediaInfo `gorm:"type:jsonb" json:"avatar_info"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
