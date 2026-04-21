package models

import (
	"time"

	"gorm.io/gorm"
)

type ZoomMeeting struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ZoomMeetingID string         `gorm:"size:255;uniqueIndex" json:"zoom_meeting_id"`
	UserID        int64          `gorm:"not null;index" json:"user_id"`
	CourseID      *int64         `gorm:"index" json:"course_id"`
	LessonID      *int64         `gorm:"index" json:"lesson_id"`
	Title         string         `gorm:"size:500" json:"title"`
	Description   string         `gorm:"type:text" json:"description"`
	StartTime     time.Time      `json:"start_time"`
	Duration      int            `json:"duration"` // minutes
	JoinURL       string         `gorm:"type:text" json:"join_url"`
	HostURL       string         `gorm:"type:text" json:"host_url"`
	Password      string         `gorm:"size:50" json:"password"`
	MeetingType   int            `gorm:"default:2" json:"meeting_type"` // 1=instant, 2=scheduled
	Status        string         `gorm:"size:50;default:scheduled" json:"status"`
	IsRecording   bool           `gorm:"default:false" json:"is_recording"`
	WaitingRoom   bool           `gorm:"default:true" json:"waiting_room"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations
	User   User    `gorm:"foreignKey:UserID"`
	Course *Course `gorm:"foreignKey:CourseID"`
	Lesson *Lesson `gorm:"foreignKey:LessonID"`
}

func (ZoomMeeting) TableName() string {
	return "zoom_meetings"
}

// Meeting status constants
const (
	MeetingStatusScheduled = "scheduled"
	MeetingStatusStarted   = "started"
	MeetingStatusEnded     = "ended"
	MeetingStatusCancelled = "cancelled"
)

// Meeting type constants
const (
	MeetingTypeInstant   = 1
	MeetingTypeScheduled = 2
	MeetingTypeRecurring = 3
)
