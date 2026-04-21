package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type NotificationData struct {
	DataValue    string `json:"value"`
	Type         string `json:"type"`
	SuccessCount int32  `json:"success_count"`
	FailedCount  int32  `json:"failed_count"`
	Error        string `json:"error"`
}

func (nd *NotificationData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan NotificationData: wrong type %T", value)
	}
	return json.Unmarshal(bytes, nd)
}

func (nd NotificationData) Value() (driver.Value, error) {
	return json.Marshal(nd)
}

type NotificationLog struct {
	ID        int64            `gorm:"primaryKey" json:"id"`
	NoticeID  int64            `gorm:"not null;index" json:"notice_id"`
	Status    string           `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	Data      NotificationData `gorm:"type:jsonb" json:"data"`
	SentAt    time.Time        `gorm:"default:CURRENT_TIMESTAMP;index" json:"sent_at"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
	Notice    *Notice          `gorm:"foreignKey:NoticeID;constraint:OnDelete:CASCADE" json:"notice,omitempty"`
}

func (NotificationLog) TableName() string {
	return "notification_logs"
}

const (
	NotificationStatusPending   = "pending"
	NotificationStatusSent      = "sent"
	NotificationStatusSuccess   = "success"
	NotificationStatusDelivered = "delivered"
	NotificationStatusFailed    = "failed"
	NotificationStatusRead      = "read"
)
