package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	HeadingMiniTime = 5
)

type Heading struct {
	ID          int64  `gorm:"primaryKey" json:"id"`
	ChapterId   int64  `gorm:"not null" json:"chapter_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Target      string `json:"target"`
	Time        int32  `json:"time"`

	Chapter      Chapter  `gorm:"foreignKey:ChapterId" json:"chapter"`
	Lessons      []Lesson `gorm:"foreignKey:HeadingID"`
	SortPosition int      `json:"sort_position"`

	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy int64          `json:"deleted_by"`
}
