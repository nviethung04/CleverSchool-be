package models

import "time"

type Teacher struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID   int64  `gorm:"not null" json:"user_id"`
	User     User   `gorm:"foreignKey:UserID"`
	Code     string `gorm:"size:20;not null;unique" json:"code"`
	School   string `gorm:"size:100;null" json:"school"`
	Position string `gorm:"size:50;null" json:"position"`

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP" json:"updated_at"`
}
