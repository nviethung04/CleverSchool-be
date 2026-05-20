package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
)

type NotificationLogRepository interface {
	CreateLog(log *models.NotificationLog) error
	GetLogsByNoticeID(noticeID int64) ([]models.NotificationLog, error)
}

type notificationLogRepository struct{}

func NewNotificationLogRepository() NotificationLogRepository {
	return &notificationLogRepository{}
}

func (r *notificationLogRepository) CreateLog(log *models.NotificationLog) error {
	return db.MasterDB.Create(log).Error
}

func (r *notificationLogRepository) GetLogsByNoticeID(noticeID int64) ([]models.NotificationLog, error) {
	var logs []models.NotificationLog
	err := db.ReplicaDB.
		Where("notice_id = ? AND deleted_at IS NULL", noticeID).
		Preload("User").
		Preload("Notice").
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

