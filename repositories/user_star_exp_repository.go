package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
)

type UserStarExpRepository interface {
	GetCurrentByUserID(userID int64) (*models.UserStarExp, error)
	Create(userStarExp *models.UserStarExp) error
	UpdateIsCurrentForUser(userID int64, isCurrent bool) error
}

type userStarExpRepository struct{}

func NewUserStarExpRepository() UserStarExpRepository {
	return &userStarExpRepository{}
}

func (r *userStarExpRepository) GetCurrentByUserID(userID int64) (*models.UserStarExp, error) {
	var userStarExp models.UserStarExp
	res := db.ReplicaDB.
		Where("user_id = ? AND is_current = ?", userID, true).
		Limit(1).
		Find(&userStarExp)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, nil // Không tìm thấy, trả về nil (chưa có record)
	}
	return &userStarExp, nil
}

func (r *userStarExpRepository) Create(userStarExp *models.UserStarExp) error {
	return db.MasterDB.Create(userStarExp).Error
}

func (r *userStarExpRepository) UpdateIsCurrentForUser(userID int64, isCurrent bool) error {
	return db.MasterDB.Model(&models.UserStarExp{}).
		Where("user_id = ?", userID).
		Update("is_current", isCurrent).Error
}
