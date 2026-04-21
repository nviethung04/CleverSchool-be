package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	MicrosoftMeetingStatusScheduled = "scheduled"
	MicrosoftMeetingStatusActive    = "active"
	MicrosoftMeetingStatusEnded     = "ended"
	MicrosoftMeetingStatusCancelled = "cancelled"
)

type MicrosoftMeeting struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	OnlineMeetingID string    `json:"online_meeting_id" gorm:"not null;uniqueIndex"` // Microsoft Graph OnlineMeeting ID
	CalendarEventID string    `json:"calendar_event_id" gorm:"uniqueIndex"`          // Optional Calendar Event ID
	UserID          uint      `json:"user_id" gorm:"not null;index"`                 // Creator (teacher)
	CourseID        *uint     `json:"course_id" gorm:"index"`                        // Optional course association
	LessonID        *uint     `json:"lesson_id" gorm:"index"`                        // Optional lesson association
	Title           string    `json:"title" gorm:"not null"`
	Description     string    `json:"description" gorm:"type:text"`
	StartTime       time.Time `json:"start_time" gorm:"not null"`
	EndTime         time.Time `json:"end_time" gorm:"not null"`
	// Local display strings (not persisted)
	StartTimeLocal    string         `json:"start_time_local" gorm:"-"`
	EndTimeLocal      string         `json:"end_time_local" gorm:"-"`
	JoinURL           string         `json:"join_url" gorm:"not null"` // Teams meeting join URL
	JoinWebURL        string         `json:"join_web_url"`             // Web-based join URL
	ConferenceID      string         `json:"conference_id"`            // Conference ID for phone dial-in
	TollNumber        string         `json:"toll_number"`              // Phone number for dial-in
	TollFreeNumber    string         `json:"toll_free_number"`         // Toll-free phone number
	TimeZone          string         `json:"time_zone" gorm:"default:'UTC'"`
	Status            string         `json:"status" gorm:"default:'scheduled'"` // scheduled, active, ended, cancelled
	IsRecurring       bool           `json:"is_recurring" gorm:"default:false"`
	RecurrenceRule    string         `json:"recurrence_rule"`            // RRULE for recurring meetings
	Attendees         string         `json:"attendees" gorm:"type:text"` // JSON array of attendee emails
	ShortCode         string         `json:"short_code" gorm:"uniqueIndex"`
	RecordingURL      string         `json:"recording_url" gorm:"type:text"`
	RecordingStatus   string         `json:"recording_status" gorm:"default:'none'"` // none, processing, available
	RecordingShareURL string         `json:"recording_share_url" gorm:"type:text"`
	ClassID           *uint          `json:"class_id" gorm:"index"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `json:"DeletedAt" gorm:"index"`

	// Relationships
	User        User                `json:"User" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Course      *Course             `json:"Course" gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Lesson      *Lesson             `json:"Lesson" gorm:"foreignKey:LessonID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Class       *Class              `json:"Class" gorm:"foreignKey:ClassID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Attendances []MeetingAttendance `json:"attendances" gorm:"foreignKey:MeetingID"`
}

func (MicrosoftMeeting) TableName() string {
	return "microsoft_meetings"
}

type MicrosoftMeetingAttendee struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	Role        string `json:"role,omitempty"` // organizer, attendee, presenter
}
