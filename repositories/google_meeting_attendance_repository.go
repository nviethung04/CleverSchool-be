package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"
)

type GoogleMeetingAttendanceRepository interface {
	Create(attendance *models.GoogleMeetingAttendance) error
	GetByID(id uint) (*models.GoogleMeetingAttendance, error)
	GetByMeetingID(meetingID uint) ([]*models.GoogleMeetingAttendance, error)
	GetByUserID(userID uint) ([]*models.GoogleMeetingAttendance, error)
	GetByMeetingAndUser(meetingID, userID uint) (*models.GoogleMeetingAttendance, error)
	Update(attendance *models.GoogleMeetingAttendance) error
	UpdateLeftAt(id uint, leftAt time.Time) error
	GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error)
	GetUserAttendanceHistory(userID uint, limit, offset int) ([]*models.GoogleMeetingAttendance, error)
	CheckUserAttended(meetingID, userID uint) (bool, error)
}

type googleMeetingAttendanceRepository struct{}

func NewGoogleMeetingAttendanceRepository() GoogleMeetingAttendanceRepository {
	return &googleMeetingAttendanceRepository{}
}

func (r *googleMeetingAttendanceRepository) Create(attendance *models.GoogleMeetingAttendance) error {
	return db.MasterDB.Create(attendance).Error
}

func (r *googleMeetingAttendanceRepository) GetByID(id uint) (*models.GoogleMeetingAttendance, error) {
	var attendance models.GoogleMeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").Preload("User").First(&attendance, id).Error
	return &attendance, err
}

func (r *googleMeetingAttendanceRepository) GetByMeetingID(meetingID uint) ([]*models.GoogleMeetingAttendance, error) {
	var attendances []*models.GoogleMeetingAttendance
	err := db.ReplicaDB.Preload("User").
		Where("meeting_id = ?", meetingID).
		Order("joined_at ASC").
		Find(&attendances).Error
	return attendances, err
}

func (r *googleMeetingAttendanceRepository) GetByUserID(userID uint) ([]*models.GoogleMeetingAttendance, error) {
	var attendances []*models.GoogleMeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Find(&attendances).Error
	return attendances, err
}

func (r *googleMeetingAttendanceRepository) GetByMeetingAndUser(meetingID, userID uint) (*models.GoogleMeetingAttendance, error) {
	var attendance models.GoogleMeetingAttendance
	err := db.ReplicaDB.
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Order("joined_at DESC").
		First(&attendance).Error
	return &attendance, err
}

func (r *googleMeetingAttendanceRepository) Update(attendance *models.GoogleMeetingAttendance) error {
	return db.MasterDB.Save(attendance).Error
}

func (r *googleMeetingAttendanceRepository) UpdateLeftAt(id uint, leftAt time.Time) error {
	return db.MasterDB.Model(&models.GoogleMeetingAttendance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"left_at":          leftAt,
			"duration_minutes": db.MasterDB.Raw("EXTRACT(EPOCH FROM (? - joined_at)) / 60", leftAt),
		}).Error
}

func (r *googleMeetingAttendanceRepository) GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error) {
	var summary models.AttendanceSummary
	summary.MeetingID = meetingID

	var totalJoined int64
	db.ReplicaDB.Model(&models.GoogleMeetingAttendance{}).
		Where("meeting_id = ?", meetingID).
		Count(&totalJoined)
	summary.TotalJoined = int(totalJoined)

	var totalPresent int64
	db.ReplicaDB.Model(&models.GoogleMeetingAttendance{}).
		Where("meeting_id = ? AND is_present = true", meetingID).
		Count(&totalPresent)
	summary.TotalPresent = int(totalPresent)

	var avgDuration float64
	db.ReplicaDB.Model(&models.GoogleMeetingAttendance{}).
		Where("meeting_id = ? AND duration_minutes > 0", meetingID).
		Select("COALESCE(AVG(duration_minutes), 0)").
		Scan(&avgDuration)
	summary.AverageDuration = int(avgDuration)

	return &summary, nil
}

func (r *googleMeetingAttendanceRepository) GetUserAttendanceHistory(userID uint, limit, offset int) ([]*models.GoogleMeetingAttendance, error) {
	var attendances []*models.GoogleMeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").Preload("Meeting.Course").Preload("Meeting.Lesson").
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Limit(limit).Offset(offset).
		Find(&attendances).Error
	return attendances, err
}

func (r *googleMeetingAttendanceRepository) CheckUserAttended(meetingID, userID uint) (bool, error) {
	var count int64
	err := db.ReplicaDB.Model(&models.GoogleMeetingAttendance{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	return count > 0, err
}
