package models

import (
	"time"

	"gorm.io/gorm"
)

type Chapter struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ProgramId    int64  `gorm:"null" json:"program_id"`
	Title        string `gorm:"size:255;not null" json:"title"`
	ObjectTitle  string `gorm:"size:255;not null" json:"object_title"`
	Description  string `gorm:"type:text;not null" json:"description"`
	Status       bool   `gorm:"null" json:"status"`
	SortPosition int    `json:"sort_position"`

	Lessons  []Lesson  `gorm:"foreignKey:ChapterID" json:"lessons"`
	Headings []Heading `gorm:"foreignKey:ChapterId" json:"headings"`
	// Course  Course   `gorm:"foreignKey:CourseId"`
	Program Program `gorm:"foreignKey:ProgramId"`

	Time   int32  `json:"time"`
	Target string `gorm:"type:text;not null" json:"target"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
