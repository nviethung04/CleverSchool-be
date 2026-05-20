package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
)

type ZoomMeetingRepository interface {
	Create(meeting *models.ZoomMeeting) error
	GetByID(id int64) (*models.ZoomMeeting, error)
	GetByZoomMeetingID(zoomMeetingID string) (*models.ZoomMeeting, error)
	GetByUserID(userID int64, limit, offset int) ([]models.ZoomMeeting, error)
	GetByCourseID(courseID int64) ([]models.ZoomMeeting, error)
	GetByLessonID(lessonID int64) ([]models.ZoomMeeting, error)
	Update(id int64, updates map[string]interface{}) error
	Delete(id int64) error
	UpdateStatus(zoomMeetingID string, status string) error
}

type zoomMeetingRepository struct{}

func NewZoomMeetingRepository() ZoomMeetingRepository {
	return &zoomMeetingRepository{}
}

func (r *zoomMeetingRepository) Create(meeting *models.ZoomMeeting) error {
	return db.MasterDB.Create(meeting).Error
}

func (r *zoomMeetingRepository) GetByID(id int64) (*models.ZoomMeeting, error) {
	var meeting models.ZoomMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("id = ?", id).First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (r *zoomMeetingRepository) GetByZoomMeetingID(zoomMeetingID string) (*models.ZoomMeeting, error) {
	var meeting models.ZoomMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").Preload("Lesson").
		Where("zoom_meeting_id = ?", zoomMeetingID).First(&meeting).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (r *zoomMeetingRepository) GetByUserID(userID int64, limit, offset int) ([]models.ZoomMeeting, error) {
	var meetings []models.ZoomMeeting
	query := db.ReplicaDB.Preload("Course").Preload("Lesson").
		Where("user_id = ?", userID).
		Order("start_time DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&meetings).Error
	return meetings, err
}

func (r *zoomMeetingRepository) GetByCourseID(courseID int64) ([]models.ZoomMeeting, error) {
	var meetings []models.ZoomMeeting
	err := db.ReplicaDB.Preload("User").Preload("Lesson").
		Where("course_id = ?", courseID).
		Order("start_time ASC").
		Find(&meetings).Error
	return meetings, err
}

func (r *zoomMeetingRepository) GetByLessonID(lessonID int64) ([]models.ZoomMeeting, error) {
	var meetings []models.ZoomMeeting
	err := db.ReplicaDB.Preload("User").Preload("Course").
		Where("lesson_id = ?", lessonID).
		Order("start_time ASC").
		Find(&meetings).Error
	return meetings, err
}

func (r *zoomMeetingRepository) Update(id int64, updates map[string]interface{}) error {
	return db.MasterDB.Model(&models.ZoomMeeting{}).
		Where("id = ?", id).Updates(updates).Error
}

func (r *zoomMeetingRepository) Delete(id int64) error {
	return db.MasterDB.Where("id = ?", id).Delete(&models.ZoomMeeting{}).Error
}

func (r *zoomMeetingRepository) UpdateStatus(zoomMeetingID string, status string) error {
	return db.MasterDB.Model(&models.ZoomMeeting{}).
		Where("zoom_meeting_id = ?", zoomMeetingID).
		Update("status", status).Error
}
