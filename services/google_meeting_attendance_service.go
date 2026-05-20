package services

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GoogleMeetingAttendanceService interface {
	JoinMeeting(meetingID, userID uint, joinMethod, deviceInfo, ipAddress string) (*models.GoogleMeetingAttendance, error)
	LeaveMeeting(meetingID, userID uint) error
	GetMeetingAttendances(meetingID uint) ([]*models.GoogleMeetingAttendance, error)
	GetUserAttendanceHistory(userID uint, page, limit int) ([]*models.GoogleMeetingAttendance, error)
	GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error)
	JoinByShortCode(shortCode string, userID uint, roleId int, deviceInfo, ipAddress string) (*models.GoogleMeeting, *models.GoogleMeetingAttendance, error)
}

type googleMeetingAttendanceService struct {
	attendanceRepo repositories.GoogleMeetingAttendanceRepository
	meetingRepo    repositories.GoogleMeetingRepository
}

func NewGoogleMeetingAttendanceService() GoogleMeetingAttendanceService {
	return &googleMeetingAttendanceService{
		attendanceRepo: repositories.NewGoogleMeetingAttendanceRepository(),
		meetingRepo:    repositories.NewGoogleMeetingRepository(),
	}
}

func (s *googleMeetingAttendanceService) isUserInCourse(userID uint, courseID uint) (bool, error) {
	var count int64
	err := db.ReplicaDB.Table("user_courses").
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *googleMeetingAttendanceService) JoinMeeting(meetingID, userID uint, joinMethod, deviceInfo, ipAddress string) (*models.GoogleMeetingAttendance, error) {
	meeting, err := s.meetingRepo.GetByID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %v", err)
	}

	if meeting.CourseID != nil {
		ok, err := s.isUserInCourse(userID, *meeting.CourseID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate course membership: %v", err)
		}
		if !ok {
			return nil, fmt.Errorf("forbidden: user is not enrolled in this course")
		}
	}

	existing, err := s.attendanceRepo.GetByMeetingAndUser(meetingID, userID)
	if err == nil && existing.LeftAt == nil {
		config.Log.Infof("User %d already in google meeting %d", userID, meetingID)
		return existing, nil
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing attendance: %v", err)
		}
	}

	attendance := &models.GoogleMeetingAttendance{
		MeetingID:  meetingID,
		UserID:     userID,
		JoinedAt:   time.Now(),
		JoinMethod: joinMethod,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IsPresent:  true,
	}

	if err := s.attendanceRepo.Create(attendance); err != nil {
		return nil, fmt.Errorf("failed to create attendance record: %v", err)
	}

	config.Log.Infof("✅ User %d joined google meeting %d (%s)", userID, meetingID, meeting.Title)
	return attendance, nil
}

func (s *googleMeetingAttendanceService) LeaveMeeting(meetingID, userID uint) error {
	attendance, err := s.attendanceRepo.GetByMeetingAndUser(meetingID, userID)
	if err != nil {
		return fmt.Errorf("attendance record not found: %v", err)
	}

	leftAt := time.Now()
	duration := int(leftAt.Sub(attendance.JoinedAt).Minutes())

	attendance.LeftAt = &leftAt
	attendance.DurationMinutes = duration

	if err := s.attendanceRepo.Update(attendance); err != nil {
		return fmt.Errorf("failed to update attendance: %v", err)
	}

	config.Log.Infof("✅ User %d left google meeting %d after %d minutes", userID, meetingID, duration)
	return nil
}

func (s *googleMeetingAttendanceService) GetMeetingAttendances(meetingID uint) ([]*models.GoogleMeetingAttendance, error) {
	return s.attendanceRepo.GetByMeetingID(meetingID)
}

func (s *googleMeetingAttendanceService) GetUserAttendanceHistory(userID uint, page, limit int) ([]*models.GoogleMeetingAttendance, error) {
	offset := (page - 1) * limit
	return s.attendanceRepo.GetUserAttendanceHistory(userID, limit, offset)
}

func (s *googleMeetingAttendanceService) GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error) {
	return s.attendanceRepo.GetAttendanceSummary(meetingID)
}

func (s *googleMeetingAttendanceService) JoinByShortCode(shortCode string, userID uint, roleId int, deviceInfo, ipAddress string) (*models.GoogleMeeting, *models.GoogleMeetingAttendance, error) {
	meeting, err := s.meetingRepo.GetByShortCode(shortCode)
	if err != nil {
		return nil, nil, fmt.Errorf("meeting not found with code: %s", shortCode)
	}

	if meeting.CourseID != nil {
		if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
			ok, err := s.isUserInCourse(userID, *meeting.CourseID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to validate course membership: %v", err)
			}
			if !ok {
				return nil, nil, fmt.Errorf("forbidden: user is not enrolled in this course")
			}
		}
	}

	now := time.Now()
	if now.Before(meeting.StartTime.Add(-15 * time.Minute)) {
		return meeting, nil, fmt.Errorf("meeting has not started yet")
	}
	if now.After(meeting.EndTime.Add(30 * time.Minute)) {
		return meeting, nil, fmt.Errorf("meeting has ended")
	}

	attendance, err := s.JoinMeeting(meeting.ID, userID, "short_link", deviceInfo, ipAddress)
	if err != nil {
		return meeting, nil, err
	}

	return meeting, attendance, nil
}
