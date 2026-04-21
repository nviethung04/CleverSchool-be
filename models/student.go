package models

import "time"

type Student struct {
	ID     int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID int64  `gorm:"not null" json:"user_id"`
	User   User   `gorm:"foreignKey:UserID"`
	School string `gorm:"size:100;null" json:"school"`
	// Courses []*Course `gorm:"many2many:class_ref_students;" json:"courses"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
}
