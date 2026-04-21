package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"time"
)

type GoogleMeetingRepository interface {
	Create(meeting *models.GoogleMeeting) error
	GetByID(id uint) (*models.GoogleMeeting, error)
	GetByCalendarEventID(calendarEventID string) (*models.GoogleMeeting, error)
	GetByShortCode(shortCode string) (*models.GoogleMeeting, error)
	GetByUserID(userID uint) ([]*models.GoogleMeeting, error)
	GetByCourseID(courseID uint) ([]*models.GoogleMeeting, error)
	GetByLessonID(lessonID uint) ([]*models.GoogleMeeting, error)
	Update(meeting *models.GoogleMeeting) error
	Delete(id uint) error
	ForceDelete(id uint) error
	LogForceDelete(meetingID uint, deletedBy uint) error
	GetUpcoming(userID uint, limit int) ([]*models.GoogleMeeting, error)
	UpdateRecording(id uint, recordingURL, status string, durationMinutes int) error
}

type googleMeetingRepository struct{}

func NewGoogleMeetingRepository() GoogleMeetingRepository {
	return &googleMeetingRepository{}
}

// Create creates a new Google meeting
func (r *googleMeetingRepository) Create(meeting *models.GoogleMeeting) error {
	return db.MasterDB.Create(meeting).Error
}

// GetByID gets meeting by ID with preloaded relationships
func (r *googleMeetingRepository) GetByID(id uint) (*models.GoogleMeeting, error) {
	var meeting models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").First(&meeting, id).Error
	return &meeting, err
}

// GetByCalendarEventID gets meeting by Google Calendar event ID
func (r *googleMeetingRepository) GetByCalendarEventID(calendarEventID string) (*models.GoogleMeeting, error) {
	var meeting models.GoogleMeeting
	err := db.ReplicaDB.Where("calendar_event_id = ?", calendarEventID).First(&meeting).Error
	return &meeting, err
}

func (r *googleMeetingRepository) GetByShortCode(shortCode string) (*models.GoogleMeeting, error) {
	var meeting models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("short_code = ?", shortCode).
		First(&meeting).Error
	return &meeting, err
}

// GetByUserID gets meetings by user ID with pagination
func (r *googleMeetingRepository) GetByUserID(userID uint) ([]*models.GoogleMeeting, error) {
	var meetings []*models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("user_id = ?", userID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

// GetByCourseID gets meetings by course ID
func (r *googleMeetingRepository) GetByCourseID(courseID uint) ([]*models.GoogleMeeting, error) {
	var meetings []*models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("course_id = ?", courseID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

// GetByLessonID gets meetings by lesson ID
func (r *googleMeetingRepository) GetByLessonID(lessonID uint) ([]*models.GoogleMeeting, error) {
	var meetings []*models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("lesson_id = ?", lessonID).
		Order("start_time DESC").
		Find(&meetings).Error
	return meetings, err
}

// Update updates meeting information
func (r *googleMeetingRepository) Update(meeting *models.GoogleMeeting) error {
	return db.MasterDB.Save(meeting).Error
}

// Delete soft deletes a meeting
func (r *googleMeetingRepository) Delete(id uint) error {
	return db.MasterDB.Delete(&models.GoogleMeeting{}, id).Error
}

func (r *googleMeetingRepository) ForceDelete(id uint) error {
	return db.MasterDB.Unscoped().Where("id = ?", id).Delete(&models.GoogleMeeting{}).Error
}

func (r *googleMeetingRepository) LogForceDelete(meetingID uint, deletedBy uint) error {
	return db.MasterDB.Exec(
		"INSERT INTO meeting_delete_logs (meeting_id, deleted_by, deleted_at) VALUES (?, ?, NOW())",
		meetingID, deletedBy,
	).Error
}

// GetUpcoming gets upcoming meetings for a user
func (r *googleMeetingRepository) GetUpcoming(userID uint, limit int) ([]*models.GoogleMeeting, error) {
	var meetings []*models.GoogleMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("user_id = ? AND start_time > ? AND status = ?", userID, time.Now(), models.GoogleMeetingStatusScheduled).
		Order("start_time ASC").
		Limit(limit).
		Find(&meetings).Error
	return meetings, err
}

func (r *googleMeetingRepository) UpdateRecording(id uint, recordingURL, status string, durationMinutes int) error {
	return db.MasterDB.Model(&models.GoogleMeeting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"recording_url":              recordingURL,
			"recording_status":           status,
			"recording_duration_minutes": durationMinutes,
		}).Error
}
