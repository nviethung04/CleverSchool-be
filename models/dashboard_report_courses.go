package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type DashboardReportCourses struct {
	ID                        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CourseID                  int64     `json:"course_id" gorm:"not null"`
	TotalStudents             int64     `json:"total_students" gorm:"default:0"`
	TotalTeachers             int64     `json:"total_teachers" gorm:"default:0"`
	ActiveStudents            int64     `json:"active_students" gorm:"default:0"`
	ActiveTeachers            int64     `json:"active_teachers" gorm:"default:0"`
	StudentsCompletedHomework int64     `json:"students_completed_homework" gorm:"default:0"`
	TotalHomeworks            int64     `json:"total_homeworks" gorm:"default:0"`           // Tổng số homework trong khoảng thời gian
	AssignedHomeworks         int64     `json:"assigned_homeworks" gorm:"default:0"`       // Số homework đã được giao (assigned_at IS NOT NULL)
	CompletedHomeworks        int64     `json:"completed_homeworks" gorm:"default:0"`      // Số homework đã hoàn thành (có ít nhất 1 user hoàn thành)
	TeacherIDs                string    `json:"teacher_ids" gorm:"type:text;default:''"`   // Danh sách ID các teacher, phân cách bằng dấu phẩy
	TeacherInfos              DashboardTeacherInfos `json:"teacher_infos" gorm:"type:jsonb;default:'[]'"` // JSON array chứa thông tin chi tiết teachers
	StartDate                 time.Time `json:"start_date" gorm:"not null"`
	EndDate                   time.Time `json:"end_date" gorm:"not null"`
	CreatedAt                 time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                 time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:ID"`
}

// DashboardTeacherInfo struct cho thông tin teacher trong JSONB
type DashboardTeacherInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// DashboardTeacherInfos type cho JSONB array
type DashboardTeacherInfos []DashboardTeacherInfo

// Value implements driver.Valuer interface
func (t DashboardTeacherInfos) Value() (driver.Value, error) {
	if t == nil {
		return "[]", nil
	}
	return json.Marshal(t)
}

// Scan implements sql.Scanner interface
func (t *DashboardTeacherInfos) Scan(value interface{}) error {
	if value == nil {
		*t = DashboardTeacherInfos{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, t)
}

func (DashboardReportCourses) TableName() string {
	return "dashboard_report_courses"
}
