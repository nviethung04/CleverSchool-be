package models

import (
	"time"

	"gorm.io/gorm"
)

// Meeting status constants
const (
	GoogleMeetingStatusScheduled = "scheduled"
	GoogleMeetingStatusActive    = "active"
	GoogleMeetingStatusEnded     = "ended"
	GoogleMeetingStatusCancelled = "cancelled"
)

// GoogleMeeting represents a Google Meet meeting created through Calendar API
type GoogleMeeting struct {
	ID                       uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	CalendarEventID          string         `json:"calendar_event_id" gorm:"not null;uniqueIndex"` // Google Calendar Event ID
	UserID                   uint           `json:"user_id" gorm:"not null;index"`                 // Creator (teacher)
	CourseID                 *uint          `json:"course_id" gorm:"index"`                        // Optional course association
	LessonID                 *uint          `json:"lesson_id" gorm:"index"`                        // Optional lesson association
	Title                    string         `json:"title" gorm:"not null"`
	Description              string         `json:"description" gorm:"type:text"`
	StartTime                time.Time      `json:"start_time" gorm:"not null"`
	EndTime                  time.Time      `json:"end_time" gorm:"not null"`
	MeetURL                  string         `json:"meet_url" gorm:"not null"` // Google Meet link
	CalendarURL              string         `json:"calendar_url"`             // Calendar event URL
	TimeZone                 string         `json:"time_zone" gorm:"default:'UTC'"`
	Status                   string         `json:"status" gorm:"default:'scheduled'"` // scheduled, active, ended, cancelled
	IsRecurring              bool           `json:"is_recurring" gorm:"default:false"`
	RecurrenceRule           string         `json:"recurrence_rule"`            // RRULE for recurring meetings
	Attendees                string         `json:"attendees" gorm:"type:text"` // JSON array of attendee emails
	ShortCode                string         `json:"short_code" gorm:"uniqueIndex"`
	RecordingURL             string         `json:"recording_url" gorm:"type:text"`
	RecordingStatus          string         `json:"recording_status" gorm:"default:'none'"` // none, processing, available
	RecordingDurationMinutes int            `json:"recording_duration_minutes" gorm:"default:0"`
	CreatedAt                time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt                gorm.DeletedAt `json:"DeletedAt" gorm:"index"`

	// Computed fields (not stored)
	JoinURL       string `json:"join_url" gorm:"-"`        // Added this line
	PublicJoinURL string `json:"public_join_url" gorm:"-"` // Added this line

	// Relationships
	User   User    `json:"User" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Course *Course `json:"Course" gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Lesson *Lesson `json:"Lesson" gorm:"foreignKey:LessonID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// TableName specifies the table name for GoogleMeeting model
func (GoogleMeeting) TableName() string {
	return "google_meetings"
}

// GoogleMeetingAttendee represents an attendee for a Google Meeting
type GoogleMeetingAttendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"` // needsAction, accepted, declined, tentative
}
