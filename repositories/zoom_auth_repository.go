package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
)

type ZoomAuthRepository interface {
	Create(zoomAccount *models.ZoomAccount) error
	GetByUserID(userID int64) (*models.ZoomAccount, error)
	UpdateTokens(userID int64, accessToken, refreshToken string, expiresAt int64) error
	Delete(userID int64) error
	IsConnected(userID int64) bool
}

type zoomAuthRepository struct{}

func NewZoomAuthRepository() ZoomAuthRepository {
	return &zoomAuthRepository{}
}

func (r *zoomAuthRepository) Create(zoomAccount *models.ZoomAccount) error {
	return db.MasterDB.Create(zoomAccount).Error
}

func (r *zoomAuthRepository) GetByUserID(userID int64) (*models.ZoomAccount, error) {
	var account models.ZoomAccount
	err := db.ReplicaDB.Where("user_id = ? AND is_active = ?", userID, true).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *zoomAuthRepository) UpdateTokens(userID int64, accessToken, refreshToken string, expiresAt int64) error {
	updates := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
	}
	return db.MasterDB.Model(&models.ZoomAccount{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Updates(updates).Error
}

func (r *zoomAuthRepository) Delete(userID int64) error {
	return db.MasterDB.Where("user_id = ?", userID).Delete(&models.ZoomAccount{}).Error
}

func (r *zoomAuthRepository) IsConnected(userID int64) bool {
	var count int64
	db.ReplicaDB.Model(&models.ZoomAccount{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count)
	return count > 0
}
