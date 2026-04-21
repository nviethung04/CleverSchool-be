package models

import (
	"time"
)

type MessageMedia struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MessageID uint64    `gorm:"not null" json:"message_id"`
	MediaID   int64     `gorm:"not null" json:"media_id"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Message ChatMessage `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	Media   Media       `gorm:"foreignKey:MediaID" json:"media,omitempty"`
}

func (MessageMedia) TableName() string {
	return "message_medias"
}
