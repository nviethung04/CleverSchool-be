package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TeacherInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

func (t TeacherInfo) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TeacherInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TeacherInfo: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

type Class struct {
	ID           int64    `json:"id" gorm:"primaryKey"`
	SchoolId     int64    `gorm:"null" json:"school_id"`
	FacultyId    int64    `gorm:"null" json:"faculty_id"`
	GradeId      int64    `gorm:"null" json:"grade_id"`
	Name         string   `json:"name"`
	Status       bool     `gorm:"null" json:"status"`
	SortPosition int      `json:"sort_position"`
	School       *School  `gorm:"foreignKey:SchoolId"`
	Grade        *Grade   `gorm:"foreignKey:GradeId"`
	Faculty      *Faculty `gorm:"foreignKey:FacultyId"`

	CurrentStudents int32 `gorm:"null" json:"current_students"`
	MaxStudents     int32 `gorm:"null" json:"max_students"`

	TeacherInfo TeacherInfo `gorm:"type:jsonb" json:"teacher_info"`

	ClassMainId int64     `gorm:"null" json:"class_main_id"`
	ClassMain   *ClassMain `gorm:"foreignKey:ClassMainId"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}

type ClassMain struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name"`
	SchoolId  int64          `gorm:"null" json:"school_id"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}

// TableName chỉ định tên bảng cho ClassMain
func (ClassMain) TableName() string {
	return "classes_main"
}
