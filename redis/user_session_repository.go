package redis

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"time"
)

type UserSessionRepository interface {
	Save(session models.UserSession) error
	DeleteByID(id string) error
	DeleteByToken(token string) error
	DeleteByUserID(id int) error
	DeleteExpiredSessions(before time.Time) error
	GetByID(id string) (*models.UserSession, error)
	GetByToken(token string) (*models.UserSession, error)
	ListByUserID(userID int) ([]models.UserSession, error)
	PermanentlyDeleteOldRecords() error
}

type userSessionRepository struct{}

func NewUserSessionRepository() UserSessionRepository {
	return &userSessionRepository{}
}

func (r *userSessionRepository) Save(session models.UserSession) error {
	return db.MasterDB.Save(&session).Error
}

func (r *userSessionRepository) DeleteByID(id string) error {
	var session models.UserSession
	return db.MasterDB.Where("id = ?", id).Delete(&session).Error
}

func (r *userSessionRepository) DeleteByUserID(id int) error {
	var session models.UserSession
	return db.MasterDB.Where("user_id = ?", id).Delete(&session).Error
}

func (r *userSessionRepository) DeleteByToken(token string) error {
	var session models.UserSession
	return db.MasterDB.Where("token = ?", token).Delete(&session).Error
}

func (r *userSessionRepository) DeleteExpiredSessions(before time.Time) error {
	var session models.UserSession
	return db.MasterDB.
		Where("expires <= ?", before).
		Delete(&session).Error
}

func (r *userSessionRepository) GetByID(id string) (*models.UserSession, error) {
	var session models.UserSession
	err := db.MasterDB.Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) GetByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := db.MasterDB.Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userSessionRepository) ListByUserID(userID int) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := db.MasterDB.Where("user_id = ?", userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *userSessionRepository) PermanentlyDeleteOldRecords() error {
	var model models.UserSession
	if !db.MasterDB.Migrator().HasTable(&model) {
		return nil
	}

	return db.MasterDB.
		Where("expires IS NOT NULL AND expires <= ?", time.Now()).
		Delete(&model).Error
}
