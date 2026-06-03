package models

import "time"

type ExamUser struct {
	ID               int64     `gorm:"primaryKey;column:id"`
	ExamID           int64     `gorm:"column:exam_id"`
	LessonID         int64     `gorm:"column:lesson_id"`
	UserID           int64     `gorm:"column:user_id"`
	Score            *float64  `gorm:"column:score"`              // numeric(5,2)
	Ratio            *float64  `gorm:"column:ratio"`              // numeric(5,2)
	Time             *int64    `gorm:"column:time"`               // bigint
	HasManualScoring bool      `gorm:"column:has_manual_scoring"` // boolean
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

// TableName sets the insert table name for this struct type
func (ExamUser) TableName() string {
	return "exam_users"
}

type ExerciseUser struct {
	ID               int64     `gorm:"primaryKey;column:id"`
	ExerciseID       int64     `gorm:"column:exercise_id"`
	LessonID         int64     `gorm:"column:lesson_id"`
	UserID           int64     `gorm:"column:user_id"`
	Score            *float64  `gorm:"column:score"`
	Ratio            *float64  `gorm:"column:ratio"`
	Time             *int64    `gorm:"column:time"`
	HasManualScoring bool      `gorm:"column:has_manual_scoring"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (ExerciseUser) TableName() string {
	return "exercise_users"
}
