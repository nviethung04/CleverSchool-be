package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID               uint64          `gorm:"primaryKey" json:"id"`
	CourseID         uint64          `gorm:"not null" json:"course_id"`
	UserID           uint64          `gorm:"not null" json:"user_id"`
	RecipientID      *uint64         `json:"recipient_id"` // NULL = group message, có giá trị = private message
	Content          *string         `json:"content"`
	MessageType      string          `gorm:"default:text" json:"message_type"`
	IsPinned         bool            `gorm:"default:false" json:"is_pinned"`
	IsEdited         bool            `gorm:"default:false" json:"is_edited"`
	EditedAt         *time.Time      `json:"edited_at"`
	ReplyToMessageID *uint64         `json:"reply_to_message_id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Relations
	Course         Course                `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User           User                  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Recipient      *User                 `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	ReplyToMessage *ChatMessage          `gorm:"foreignKey:ReplyToMessageID" json:"reply_to_message,omitempty"`
	Replies        []ChatMessage         `gorm:"foreignKey:ReplyToMessageID" json:"replies,omitempty"`
	Medias        []Media            `gorm:"many2many:message_medias;foreignKey:ID;joinForeignKey:MessageID;References:ID;joinReferences:MediaID" json:"medias,omitempty"`
	MessageMedias []MessageMedia     `gorm:"foreignKey:MessageID" json:"message_medias,omitempty"`
	Reads         []ChatMessageRead  `gorm:"foreignKey:MessageID" json:"reads,omitempty"`
}

type ChatMessageRead struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	MessageID uint64    `gorm:"not null" json:"message_id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	ReadAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"read_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
