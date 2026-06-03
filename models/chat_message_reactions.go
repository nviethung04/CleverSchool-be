package models

import (
	"time"
)

type ChatMessageReaction struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MessageID uint64    `gorm:"not null" json:"message_id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	Emoji     string    `gorm:"type:varchar(10);not null" json:"emoji"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	
	// Relations
	Message   ChatMessage `gorm:"foreignKey:MessageID" json:"message,omitempty"`
	User      User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Add unique constraint for message_id, user_id, emoji
func (ChatMessageReaction) TableName() string {
	return "chat_message_reactions"
}
