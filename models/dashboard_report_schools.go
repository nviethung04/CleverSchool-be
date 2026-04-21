package models

import (
	"time"
)

type DashboardReportSchoolWeeks struct {
	ID                        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SchoolID                  int64     `json:"school_id" gorm:"not null"`
	TotalStudents             int64     `json:"total_students" gorm:"default:0"`
	TotalTeachers             int64     `json:"total_teachers" gorm:"default:0"`
	ActiveStudents            int64     `json:"active_students" gorm:"default:0"`
	ActiveTeachers            int64     `json:"active_teachers" gorm:"default:0"`
	StudentsCompletedHomework int64     `json:"students_completed_homework" gorm:"default:0"`
	StartDate                 time.Time `json:"start_date" gorm:"not null"`
	EndDate                   time.Time `json:"end_date" gorm:"not null"`
	CreatedAt                 time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	School School `json:"school,omitempty" gorm:"foreignKey:SchoolID;references:ID"`
}

func (DashboardReportSchoolWeeks) TableName() string {
	return "dashboard_report_schools"
}
