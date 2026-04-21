package models

import (
	"time"

	"gorm.io/gorm"
)

type GoogleMeetingAttendance struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	MeetingID       uint           `json:"meeting_id" gorm:"not null;index"`
	UserID          uint           `json:"user_id" gorm:"not null;index"`
	JoinedAt        time.Time      `json:"joined_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	LeftAt          *time.Time     `json:"left_at"`
	DurationMinutes int            `json:"duration_minutes" gorm:"default:0"`
	JoinMethod      string         `json:"join_method" gorm:"default:'link'"` // link, app, phone
	DeviceInfo      string         `json:"device_info"`
	IPAddress       string         `json:"ip_address"`
	IsPresent       bool           `json:"is_present" gorm:"default:true"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"DeletedAt" gorm:"index"`

	Meeting GoogleMeeting `json:"meeting" gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	User    User          `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (GoogleMeetingAttendance) TableName() string {
	return "google_meeting_attendances"
}
