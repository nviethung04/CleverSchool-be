package models

type Permission struct {
	ID           int64  `gorm:"primaryKey;autoIncrement" json:"id"` // BIGINT UNSIGNED
	Group        string `gorm:"size:100;not null" json:"group"`
	Name         string `gorm:"size:100;not null" json:"name"`
	Description  string `gorm:"size:255" json:"description"`
	Permission   string `gorm:"size:255" json:"permission"`
	SortPosition int    `gorm:"size:255" json:"sort_position"`
	IsDisplay    *bool  `gorm:"default:true" json:"is_display"`
	Roles        []Role `gorm:"many2many:role_permissions"` // Quan hệ many-to-many ngược lại
}
