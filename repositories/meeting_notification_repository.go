package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"time"
)

type MeetingNotificationRepository interface {
	BulkCreate(notifs []models.MeetingNotification) error
	ListByUser(userID uint, limit, offset int) ([]models.MeetingNotification, error)
	ListByUserAndProvider(userID uint, provider string, limit, offset int) ([]models.MeetingNotification, error)
	CountUnread(userID uint) (int64, error)
	MarkRead(id uint, userID uint) error
}

type meetingNotificationRepository struct{}

func NewMeetingNotificationRepository() MeetingNotificationRepository {
	return &meetingNotificationRepository{}
}

func (r *meetingNotificationRepository) BulkCreate(notifs []models.MeetingNotification) error {
	if len(notifs) == 0 {
		return nil
	}
	return db.MasterDB.Create(&notifs).Error
}

func (r *meetingNotificationRepository) ListByUser(userID uint, limit, offset int) ([]models.MeetingNotification, error) {
	var notifs []models.MeetingNotification
	q := db.ReplicaDB.Where("user_id = ?", userID).
		Order("read_at IS NULL DESC, created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

// ListByUserAndProvider returns notifications filtered by provider.
// provider: "google" or "microsoft"; if empty, returns all.
func (r *meetingNotificationRepository) ListByUserAndProvider(userID uint, provider string, limit, offset int) ([]models.MeetingNotification, error) {
	var notifs []models.MeetingNotification

	q := db.ReplicaDB.Model(&models.MeetingNotification{}).
		Where("user_id = ?", userID)

	if provider != "" {
		q = q.Where("provider = ?", provider)
	}

	q = q.Order("read_at IS NULL DESC, created_at DESC")

	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	if err := q.Find(&notifs).Error; err != nil {
		return nil, err
	}
	return notifs, nil
}

func (r *meetingNotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := db.ReplicaDB.Model(&models.MeetingNotification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *meetingNotificationRepository) MarkRead(id uint, userID uint) error {
	return db.MasterDB.Model(&models.MeetingNotification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"read_at": time.Now(),
		}).Error
}
