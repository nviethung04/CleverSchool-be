package models

import (
	"time"

	"gorm.io/gorm"
)

type Course struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SubjectId       int64     `gorm:"null" json:"subject_id"`
	ProgramId       int64     `gorm:"null" json:"program_id"`
	Name            string    `gorm:"size:255;not null" json:"name"`
	ObjectTitle            string    `gorm:"size:255;not null" json:"object_title"`
	Description     string    `gorm:"type:text;not null" json:"description"`
	Duration        int       `json:"duration"`
	Type            string    `gorm:"size:255;not null" json:"type"`
	ImageInfo       MediaInfo `gorm:"type:jsonb" json:"image_info"`
	Level           string    `gorm:"size:255;not null" json:"level"`
	Status          bool      `gorm:"null" json:"status"`
	CurrentStudents int32     `gorm:"null" json:"current_students"`
	Target          string    `gorm:"null" json:"target"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Time            string    `gorm:"null" json:"time"`

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy int64          `gorm:"null" json:"created_by"`
	UpdatedBy int64          `gorm:"null" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	// Chapters    []Chapter    `gorm:"foreignKey:CourseId" json:"chapters"`
	Subject     Subject      `gorm:"foreignKey:SubjectId"`
	Semesters   []Semester   `gorm:"many2many:course_ref_semesters" json:"semesters"`
	Schools     []School     `gorm:"many2many:course_schools"`
	UserCourses []UserCourse `gorm:"foreignKey:CourseId"`
	Users       []User       `gorm:"many2many:user_courses"`
	Program     Program      `gorm:"foreignKey:ProgramId"`

	State string `gorm:"type:COURSE_STATE;not null"`

	CourseRefStudyShifts []CourseRefStudyShift `gorm:"foreignKey:CourseID"`
}
