package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"fmt"
	"strings"
	"time"
)

type MicrosoftMeetingRepository interface {
	Create(meeting *models.MicrosoftMeeting) error
	CreateOrGetByOnlineMeetingID(meeting *models.MicrosoftMeeting) (*models.MicrosoftMeeting, error)
	GetByID(id uint) (*models.MicrosoftMeeting, error)
	GetByOnlineMeetingID(meetingID string) (*models.MicrosoftMeeting, error)
	GetByShortCode(shortCode string) (*models.MicrosoftMeeting, error)
	GetByUserID(userID uint, limit, offset int) ([]*models.MicrosoftMeeting, error)
	GetByCourseID(courseID uint) ([]*models.MicrosoftMeeting, error)
	GetByLessonID(lessonID uint) ([]*models.MicrosoftMeeting, error)
	GetByClassID(classID uint) ([]*models.MicrosoftMeeting, error)
	Update(meeting *models.MicrosoftMeeting) error
	Delete(id uint) error
	GetUpcoming(userID uint, limit int) ([]*models.MicrosoftMeeting, error)
	UpdateRecording(id uint, recordingURL, status string) error
	UpdateRecordingShareURL(id uint, shareURL string) error
}

type microsoftMeetingRepository struct{}

func NewMicrosoftMeetingRepository() MicrosoftMeetingRepository {
	return &microsoftMeetingRepository{}
}

// Create creates a new Microsoft meeting
func (r *microsoftMeetingRepository) Create(meeting *models.MicrosoftMeeting) error {
	return db.MasterDB.Create(meeting).Error
}

// CreateOrGetByOnlineMeetingID creates a new meeting or returns existing one with same OnlineMeetingID
// Handles race conditions by trying to create and gracefully handling duplicates
func (r *microsoftMeetingRepository) CreateOrGetByOnlineMeetingID(meeting *models.MicrosoftMeeting) (*models.MicrosoftMeeting, error) {
	// Try to create the meeting using GORM (handles timestamps and hooks)
	err := db.MasterDB.Create(meeting).Error

	// If creation succeeded, return the created meeting
	if err == nil {
		// Fetch with preloads
		return r.GetByID(meeting.ID)
	}

	// If creation failed with duplicate key error, try to fetch the existing record
	// First, try a simple query without preloads to confirm the record exists
	var existingMeeting models.MicrosoftMeeting
	queryErr := db.ReplicaDB.Where("online_meeting_id = ?", meeting.OnlineMeetingID).First(&existingMeeting).Error

	if queryErr == nil {
		// Record exists, fetch it with preloads
		return r.GetByID(existingMeeting.ID)
	}

	// If we can't find the record and creation failed, return the original error
	return nil, fmt.Errorf("failed to save meeting to database: %v", err)
}

// GetByID gets meeting by ID with preloaded relationships
func (r *microsoftMeetingRepository) GetByID(id uint) (*models.MicrosoftMeeting, error) {
	var meeting models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").First(&meeting, id).Error
	return &meeting, err
}

// GetByOnlineMeetingID gets meeting by Microsoft Graph OnlineMeeting ID
func (r *microsoftMeetingRepository) GetByOnlineMeetingID(meetingID string) (*models.MicrosoftMeeting, error) {
	var meeting models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("online_meeting_id = ?", meetingID).First(&meeting).Error
	return &meeting, err
}

// GetByUserID gets meetings by user ID with pagination
func (r *microsoftMeetingRepository) GetByUserID(userID uint, limit, offset int) ([]*models.MicrosoftMeeting, error) {
	var meetings []*models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("user_id = ?", userID).
		Order("start_time DESC").
		Limit(limit).Offset(offset).
		Find(&meetings).Error
	return meetings, err
}

// GetByCourseID gets meetings by course ID
func (r *microsoftMeetingRepository) GetByCourseID(courseID uint) ([]*models.MicrosoftMeeting, error) {
	var meetings []*models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("course_id = ?", courseID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

// GetByLessonID gets meetings by lesson ID
func (r *microsoftMeetingRepository) GetByLessonID(lessonID uint) ([]*models.MicrosoftMeeting, error) {
	var meetings []*models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("lesson_id = ?", lessonID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

// Update updates meeting information
func (r *microsoftMeetingRepository) Update(meeting *models.MicrosoftMeeting) error {
	return db.MasterDB.Save(meeting).Error
}

// Delete hard deletes a meeting (permanently removes from database)
func (r *microsoftMeetingRepository) Delete(id uint) error {
	return db.MasterDB.Unscoped().Delete(&models.MicrosoftMeeting{}, id).Error
}

// GetUpcoming gets upcoming meetings for a user
func (r *microsoftMeetingRepository) GetUpcoming(userID uint, limit int) ([]*models.MicrosoftMeeting, error) {
	var meetings []*models.MicrosoftMeeting

	var courseIDs []uint
	_ = db.ReplicaDB.
		Table("user_courses").
		Where("user_id = ?", userID).
		Pluck("course_id", &courseIDs).Error

	query := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson")

	query = query.Where("start_time > ? AND status = ?", time.Now(), models.MicrosoftMeetingStatusScheduled)

	query = query.Where(
		"(user_id = ?) OR (course_id IN (?)) OR (lesson_id IN (SELECT id FROM lessons WHERE course_id IN (?)))",
		userID, courseIDs, courseIDs,
	)

	err := query.
		Distinct("microsoft_meetings.*").
		Order("start_time ASC").
		Limit(limit).
		Find(&meetings).Error
	return meetings, err
}

func (r *microsoftMeetingRepository) GetByShortCode(shortCode string) (*models.MicrosoftMeeting, error) {
	var meeting models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").Preload("Class").
		Where("short_code = ?", shortCode).First(&meeting).Error
	return &meeting, err
}

func (r *microsoftMeetingRepository) GetByClassID(classID uint) ([]*models.MicrosoftMeeting, error) {
	var meetings []*models.MicrosoftMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").Preload("Attendances").
		Where("class_id = ?", classID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

func (r *microsoftMeetingRepository) UpdateRecording(id uint, recordingURL, status string) error {
	return db.MasterDB.Model(&models.MicrosoftMeeting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"recording_url": recordingURL,
			"updated_at":    time.Now(),
		}).Error
}

func (r *microsoftMeetingRepository) UpdateRecordingShareURL(id uint, shareURL string) error {
	status := "available"
	if strings.TrimSpace(shareURL) == "" {
		status = "none"
		shareURL = ""
	}
	return db.MasterDB.Model(&models.MicrosoftMeeting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"recording_share_url": shareURL,
			"recording_status":    status,
			"updated_at":          time.Now(),
		}).Error
}
