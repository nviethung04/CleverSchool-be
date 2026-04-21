package dto

import (
	"time"
)

// MeetingAttendanceResponse represents full attendance information including user details
type MeetingAttendanceResponse struct {
	ID              uint       `json:"id"`
	MeetingID       uint       `json:"meeting_id"`
	UserID          uint       `json:"user_id"`
	Email           string     `json:"email"`        // User.Email
	Username        string     `json:"username"`     // User.Username
	FullName        string     `json:"full_name"`    // User.Name
	SchoolName      string     `json:"school_name"`  // User.School.Name
	SchoolID        int        `json:"school_id"`    // User.SchoolID
	ClassNames      []string   `json:"class_names"`  // User.UserClasses[*].Class.Name
	CourseNames     []string   `json:"course_names"` // Courses from User.UserCourses
	LessonName      string     `json:"lesson_name"`  // Meeting.Lesson.Name if exists
	CourseName      string     `json:"course_name"`  // Meeting.Course.Name if exists
	JoinedAt        time.Time  `json:"joined_at"`
	LeftAt          *time.Time `json:"left_at"`
	DurationMinutes int        `json:"duration_minutes"`
	JoinMethod      string     `json:"join_method"`
	DeviceInfo      string     `json:"device_info"`
	IPAddress       string     `json:"ip_address"`
	IsPresent       bool       `json:"is_present"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
