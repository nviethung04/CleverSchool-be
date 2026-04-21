package models

type RolePermission struct {
	RoleID       int64 `gorm:"primaryKey" json:"role_id"`
	PermissionID int64 `gorm:"primaryKey" json:"permission_id"`

	Role       Role       `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE;"`       // Liên kết với bảng roles
	Permission Permission `gorm:"foreignKey:PermissionID;constraint:OnDelete:CASCADE;"` // Liên kết với bảng permissions
}
