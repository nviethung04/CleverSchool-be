package models

import (
	"time"
)

type MeetingAttendance struct {
	ID              uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	MeetingID       uint       `json:"meeting_id" gorm:"not null;index"`
	UserID          uint       `json:"user_id" gorm:"not null;index"`
	JoinedAt        time.Time  `json:"joined_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	LeftAt          *time.Time `json:"left_at"`
	DurationMinutes int        `json:"duration_minutes" gorm:"default:0"`
	JoinMethod      string     `json:"join_method" gorm:"default:'link'"` // link, app, phone
	DeviceInfo      string     `json:"device_info"`
	IPAddress       string     `json:"ip_address"`
	IsPresent       bool       `json:"is_present" gorm:"default:true"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	Meeting MicrosoftMeeting `json:"meeting" gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	User    User             `json:"user" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (MeetingAttendance) TableName() string {
	return "meeting_attendances"
}

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLate    AttendanceStatus = "late"
	AttendanceStatusLeft    AttendanceStatus = "left_early"
)

type AttendanceSummary struct {
	MeetingID       uint   `json:"meeting_id"`
	TotalInvited    int    `json:"total_invited"`
	TotalJoined     int    `json:"total_joined"`
	TotalPresent    int    `json:"total_present"`
	TotalAbsent     int    `json:"total_absent"`
	AverageDuration int    `json:"average_duration_minutes"`
	AttendanceRate  string `json:"attendance_rate"`
}
