package models

import (
	"time"

	"gorm.io/gorm"
)

type LevelTest struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	Name            string         `json:"name"`
	CourseID        int64          `json:"course_id"`
	LevelStandardID int64          `json:"level_standard_id"`
	Status          int16          `json:"status"`
	TimeLimit       int64          `json:"time_limit"`
	MaxScore        float64        `json:"max_score"`
	Description     string         `json:"description"`
	CoverImageInfo  MediaInfo      `gorm:"type:jsonb" json:"cover_image_info"`
	Deadline        time.Time      `json:"deadline"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CreatedBy       int64          `json:"created_by"`
	UpdatedBy       int64          `json:"updated_by"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
	DeletedBy       int64          `gorm:"column:deleted_by" json:"deleted_by"`
}

type LevelStandard struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Name      int64          `json:"name"` // kiểu bigint trong DB
	CreatedBy int64          `json:"created_by" gorm:"column:create_by"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:create_at"`
	UpdatedBy int64          `json:"updated_by" gorm:"column:update_by"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:update_at"`
	DeletedBy int64          `json:"deleted_by" gorm:"column:deleted_by"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index"`
}

type Level struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	LevelStandardID int64          `json:"level_standard_id" gorm:"column:level_standard_id"`
	Name            int64          `json:"name"` // kiểu bigint
	MinPoint        float64        `json:"min_point" gorm:"type:numeric(5,2)"`
	MaxPoint        float64        `json:"max_point" gorm:"type:numeric(5,2)"`
	CreatedBy       int64          `json:"created_by" gorm:"column:created_by"`
	CreatedAt       time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedBy       int64          `json:"updated_by" gorm:"column:updated_by"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedBy       int64          `json:"deleted_by" gorm:"column:deleted_by"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index"`
}

type LevelUser struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `json:"user_id" gorm:"column:user_id"`
	LevelTestID int64     `json:"level_test_id" gorm:"column:level_test_id"`
	Score       float64   `json:"score" gorm:"type:numeric(5,2)"`
	LevelID     int64     `json:"level_id" gorm:"column:level_id"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}
