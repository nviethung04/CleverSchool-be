package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"time"
)

type MicrosoftAuthRepository interface {
	Create(account *models.MicrosoftAccount) error
	GetByUserID(userID uint) (*models.MicrosoftAccount, error)
	UpdateTokens(userID uint, accessToken, refreshToken string, expiresAt time.Time) error
	UpdateTokensWithStatus(userID uint, accessToken, refreshToken string, expiresAt time.Time, isActive bool) error
	Delete(userID uint) error
	IsConnected(userID uint) bool
}

type microsoftAuthRepository struct{}

func NewMicrosoftAuthRepository() MicrosoftAuthRepository {
	return &microsoftAuthRepository{}
}

// Create creates a new Microsoft account connection
func (r *microsoftAuthRepository) Create(account *models.MicrosoftAccount) error {
	return db.MasterDB.Create(account).Error
}

// GetByUserID gets Microsoft account by user ID (any status)
// Use master DB to avoid replica lag (token may appear empty/stale on replica)
func (r *microsoftAuthRepository) GetByUserID(userID uint) (*models.MicrosoftAccount, error) {
	var account models.MicrosoftAccount
	err := db.ReplicaDB.Preload("User").Where("user_id = ?", userID).First(&account).Error
	return &account, err
}

// UpdateTokens updates access and refresh tokens
func (r *microsoftAuthRepository) UpdateTokens(userID uint, accessToken, refreshToken string, expiresAt time.Time) error {
	return db.MasterDB.Model(&models.MicrosoftAccount{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"updated_at":    time.Now(),
		}).Error
}

func (r *microsoftAuthRepository) UpdateTokensWithStatus(userID uint, accessToken, refreshToken string, expiresAt time.Time, isActive bool) error {
	return db.MasterDB.Model(&models.MicrosoftAccount{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"is_active":     isActive,
			"updated_at":    time.Now(),
		}).Error
}

func (r *microsoftAuthRepository) UpdateTokensWithStatusAndTenantID(userID uint, accessToken, refreshToken string, expiresAt time.Time, tenantID string, isActive bool) error {
	updates := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
		"is_active":     isActive,
		"updated_at":    time.Now(),
	}
	if tenantID != "" {
		updates["tenant_id"] = tenantID
	}

	return db.MasterDB.Model(&models.MicrosoftAccount{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// UpdateTokensWithTenantID updates access/refresh tokens and tenant_id
// (without changing active status)
func (r *microsoftAuthRepository) UpdateTokensWithTenantID(userID uint, accessToken, refreshToken string, expiresAt time.Time, tenantID string) error {
	updates := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
		"updated_at":    time.Now(),
	}
	if tenantID != "" {
		updates["tenant_id"] = tenantID
	}

	return db.MasterDB.Model(&models.MicrosoftAccount{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// Delete soft deletes Microsoft account connection
func (r *microsoftAuthRepository) Delete(userID uint) error {
	return db.MasterDB.Where("user_id = ?", userID).Delete(&models.MicrosoftAccount{}).Error
}

// IsConnected checks if user has an active Microsoft account connection
func (r *microsoftAuthRepository) IsConnected(userID uint) bool {
	var count int64
	db.ReplicaDB.Model(&models.MicrosoftAccount{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count)
	return count > 0
}
