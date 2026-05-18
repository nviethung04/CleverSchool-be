package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"
)

type GoogleAuthRepository interface {
	Create(account *models.GoogleAccount) error
	GetByUserID(userID uint) (*models.GoogleAccount, error)
	UpdateTokens(userID uint, accessToken, refreshToken string, expiresAt time.Time) error
	Delete(userID uint) error
	DeactivateAndClear(userID uint) error
	IsConnected(userID uint) bool
	DeleteOtherActive(userID uint, keepID uint) error
}

type googleAuthRepository struct{}

func NewGoogleAuthRepository() GoogleAuthRepository {
	return &googleAuthRepository{}
}

// Create creates a new Google account connection
func (r *googleAuthRepository) Create(account *models.GoogleAccount) error {
	return db.MasterDB.Create(account).Error
}

// GetByUserID gets Google account by user ID (use master to avoid replica lag creating duplicates)
func (r *googleAuthRepository) GetByUserID(userID uint) (*models.GoogleAccount, error) {
	var account models.GoogleAccount
	err := db.ReplicaDB.Preload("User").
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("id DESC").
		First(&account).Error
	return &account, err
}

// UpdateTokens updates access and refresh tokens
func (r *googleAuthRepository) UpdateTokens(userID uint, accessToken, refreshToken string, expiresAt time.Time) error {
	return db.MasterDB.Model(&models.GoogleAccount{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"updated_at":    time.Now(),
		}).Error
}

// Delete soft deletes Google account connection
func (r *googleAuthRepository) Delete(userID uint) error {
	return db.MasterDB.Where("user_id = ?", userID).Delete(&models.GoogleAccount{}).Error
}

// DeactivateAndClear sets inactive and wipes tokens (keeps row for audit)
func (r *googleAuthRepository) DeactivateAndClear(userID uint) error {
	return db.MasterDB.Model(&models.GoogleAccount{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_active":     false,
			"access_token":  "",
			"refresh_token": "",
			"expires_at":    time.Time{},
			"updated_at":    time.Now(),
		}).Error
}

// DeleteOtherActive removes other active accounts for the same user to avoid duplicates
func (r *googleAuthRepository) DeleteOtherActive(userID uint, keepID uint) error {
	return db.MasterDB.
		Where("user_id = ? AND id <> ? AND is_active = ?", userID, keepID, true).
		Delete(&models.GoogleAccount{}).Error
}

// IsConnected checks if user has an active Google account connection
func (r *googleAuthRepository) IsConnected(userID uint) bool {
	var count int64
	db.ReplicaDB.Model(&models.GoogleAccount{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count)
	return count > 0
}
