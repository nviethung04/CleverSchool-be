package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64              `gorm:"primaryKey"`
	SchoolID     int                `json:"school_id"`
	Identifier   string             `json:"identifier"`
	Code         string             `json:"code"`
	Username     string             `json:"username"`
	Name         string             `json:"name"`
	Token        string             `json:"token"`
	ExpiresTime  time.Time          `json:"expires_time"`
	Password     string             `json:"password"`
	Email        string             `json:"email"`
	PhoneNumber  string             `json:"phone_number"`
	Address      string             `json:"address"`
	Status       bool               `json:"status"`
	Description  string             `json:"description"`
	AvatarInfo   MediaInfo          `gorm:"type:jsonb" json:"avatar_info"`
	DateOfBirth  time.Time          `json:"date_of_birth"`
	ParentID     int                `json:"parent_id"`
	// Role         Role               `gorm:"foreignKey:RoleID"`
	School       School             `gorm:"foreignKey:SchoolID"`
	UserClasses  []UserClass        `gorm:"foreignKey:UserId"`
	UserCourses  []UserCourse       `gorm:"foreignKey:UserId"`
	UserAddress  UserAddress        `gorm:"foreignKey:UserId;references:ID"`
	Certificates []Certificate      `gorm:"foreignKey:UserId;references:ID"`
	Degrees      []Degree           `gorm:"foreignKey:UserId;references:ID"`
	Departments  []Department       `gorm:"many2many:user_departments;joinForeignKey:UserId;joinReferences:DepartmentId"`
	Positions    []EmployeePosition `gorm:"many2many:user_positions;joinForeignKey:UserId;joinReferences:EmployeePositionId"`
	Courses      []Course           `gorm:"many2many:user_courses"`
	Classes      []Class            `gorm:"many2many:user_classes"`
	Subjects     []Subject          `gorm:"many2many:user_ref_subjects"`
	Roles     []Role          `gorm:"many2many:user_ref_roles"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   int64          `json:"created_by"`
	UpdatedBy   int64          `json:"updated_by"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy   int64          `gorm:"column:deleted_by"`
	LastLoginAt *time.Time     `gorm:"column:last_login_at" json:"last_login_at"`
	TypeTeacher int16          `gorm:"column:type_teacher;default:0" json:"type_teacher"`
}
