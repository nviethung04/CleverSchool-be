package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	AdminRoleId   = 1
	TeacherRoleId = 2
	StudentRoleId = 3
	SchoolRoleId  = 4
	ReadOnlyRoleId  = 5
	PageAdmin = "admin"
	PageTeacher = "teacher"
	PageStudent = "student"
)

type Role struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID    int64        `gorm:"column:parent_id" json:"parent_id"`
	Name        string       `gorm:"size:50;not null" json:"name"`
	DefaultPageView        string       `gorm:"size:50;not null" json:"default_page_view"`
	Users []User  `gorm:"many2many:user_ref_roles;joinForeignKey:RoleId;joinReferences:UserId"`
	Permissions []Permission `gorm:"many2many:role_permissions"`
	Status      bool         `gorm:"not null" json:"status"`
	DefaultPageID    int64        `gorm:"column:default_page_id" json:"default_page_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
