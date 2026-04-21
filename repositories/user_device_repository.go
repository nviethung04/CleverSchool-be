package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
)

type UserDeviceRepository interface {
	FindByUserID(userID int64) ([]models.UserDevice, error)
	FindByDeviceToken(token string, isActive bool) (*models.UserDevice, error)
	UpsertDevice(userID int64, deviceToken, deviceType, platform, appVersion string) error
	DeactivateDevice(deviceToken string) error
	DeactivateUserDevices(userID int64) error
	GetActiveTokensByUserIDs(userIDs []int64) ([]string, error)
	CleanupInactiveDevices() error
	GetCountDevice() int64
	GetCountDeviceByUserIds(userIDs []int64) int64
}

type userDeviceRepository struct{}

func NewUserDeviceRepository() UserDeviceRepository {
	return &userDeviceRepository{}
}

func (r *userDeviceRepository) FindByUserID(userID int64) ([]models.UserDevice, error) {
	var devices []models.UserDevice
	err := db.ReplicaDB.Where("user_id = ? AND is_active = ?", userID, true).Find(&devices).Error
	return devices, err
}

func (r *userDeviceRepository) FindByDeviceToken(token string, isActive bool) (*models.UserDevice, error) {
	var device models.UserDevice

	query := db.ReplicaDB.Where("device_token = ?", token)
	if isActive {
		query = query.Where("is_active = ?", true)
	}

	err := query.First(&device).Error
	if err != nil {
		return nil, err
	}

	return &device, nil
}

func (r *userDeviceRepository) UpsertDevice(userID int64, deviceToken, deviceType, platform, appVersion string) error {
	var device models.UserDevice

	err := db.ReplicaDB.Where("device_token = ?", deviceToken).First(&device).Error

	if err != nil {
		device = models.UserDevice{
			UserID:      userID,
			DeviceToken: deviceToken,
			DeviceType:  deviceType,
			Platform:    &platform,
			AppVersion:  &appVersion,
			IsActive:    true,
		}
		return db.MasterDB.Create(&device).Error
	}

	updates := map[string]interface{}{
		"user_id":     userID,
		"device_type": deviceType,
		"platform":    platform,
		"app_version": appVersion,
		"is_active":   true,
	}
	return db.MasterDB.Model(&device).Updates(updates).Error
}

func (r *userDeviceRepository) DeactivateDevice(deviceToken string) error {
	return db.MasterDB.Where("device_token = ?", deviceToken).Delete(&models.UserDevice{}).Error
}

func (r *userDeviceRepository) DeactivateUserDevices(userID int64) error {
	return db.MasterDB.Model(&models.UserDevice{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

func (r *userDeviceRepository) GetActiveTokensByUserIDs(userIDs []int64) ([]string, error) {
	var tokens []string
	err := db.ReplicaDB.Model(&models.UserDevice{}).
		Where("user_id IN ? AND is_active = ?", userIDs, true).
		Pluck("device_token", &tokens).Error
	return tokens, err
}

func (r *userDeviceRepository) CleanupInactiveDevices() error {
	return db.MasterDB.Where("is_active = ? AND updated_at < NOW() - INTERVAL '90 days'", false).
		Delete(&models.UserDevice{}).Error
}

func (r *userDeviceRepository) GetCountDeviceByUserIds(userIDs []int64) int64 {
	var userCount int64
	if len(userIDs) > 0 {
		db.ReplicaDB.Model(&models.UserDevice{}).
			Where("user_id IN ? AND is_active = true", userIDs).
			Distinct("user_id").
			Count(&userCount)
	}

	return userCount
}

func (r *userDeviceRepository) GetCountDevice() int64 {
	var userCount int64
	db.ReplicaDB.Model(&models.UserDevice{}).
		Where("is_active = true").
		Distinct("user_id").
		Count(&userCount)

	return userCount
}
