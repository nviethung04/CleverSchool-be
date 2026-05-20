package repositories

import (
	"be-cleverschool/database/db"
	"time"
)

type ClassUserRelationRepository interface {
	UpdateUserClassRelation(classID int64, studentIDs []int64, updatedBy int64) error
}

type classUserRelationRepository struct{}

func NewClassUserRelationRepository() ClassUserRelationRepository {
	return &classUserRelationRepository{}
}

func (r *classUserRelationRepository) UpdateUserClassRelation(classID int64, studentIDs []int64, updatedBy int64) error {
	now := time.Now()
	
	// Bắt đầu transaction
	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Bước 1: Set is_current = false cho tất cả user_ids trong danh sách
	err := tx.Table("user_classes").
		Where("user_id IN ?", studentIDs).
		Updates(map[string]interface{}{
			"is_current": false,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Bước 2: Update các bản ghi đã tồn tại với class_id này
	err = tx.Table("user_classes").
		Where("user_id IN ? AND class_id = ?", studentIDs, classID).
		Updates(map[string]interface{}{
			"is_current": true,
			"start_time": now,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Bước 3: Lấy danh sách user_ids đã có bản ghi với class_id này
	var existingUserIDs []int64
	err = tx.Table("user_classes").
		Where("user_id IN ? AND class_id = ?", studentIDs, classID).
		Pluck("user_id", &existingUserIDs).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Bước 4: Tạo map để kiểm tra nhanh
	existingMap := make(map[int64]bool)
	for _, id := range existingUserIDs {
		existingMap[id] = true
	}
	
	// Bước 5: Insert các bản ghi mới cho user_ids chưa có
	var newRecords []map[string]interface{}
	for _, userID := range studentIDs {
		if !existingMap[userID] {
			newRecords = append(newRecords, map[string]interface{}{
				"user_id":    userID,
				"class_id":   classID,
				"start_time": now,
				"is_current": true,
			})
		}
	}
	
	// Insert batch nếu có bản ghi mới
	if len(newRecords) > 0 {
		err = tx.Table("user_classes").Create(&newRecords).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	
	// Commit transaction
	return tx.Commit().Error
} 
