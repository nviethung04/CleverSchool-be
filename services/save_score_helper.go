package services

import (
	"be-Clever School/database/db"
	"gorm.io/gorm"
)

// CheckAndCleanExistingAnswers kiểm tra và xóa các câu trả lời cũ nếu cần
// table: tên bảng cần kiểm tra
// homeworkID, userID, questionID: thông tin để filter
// Returns: true nếu cần lưu câu trả lời mới, false nếu đã có câu trả lời đúng hết
func CheckAndCleanExistingAnswers(table string, homeworkID, userID, questionID int64) (bool, error) {
	// Kiểm tra xem đã có câu trả lời chưa
	var count int64
	err := db.MasterDB.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	// Nếu chưa có câu trả lời nào, cần lưu mới
	if count == 0 {
		return true, nil
	}

	// Kiểm tra xem có câu trả lời sai không
	var incorrectCount int64
	err = db.MasterDB.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false", homeworkID, userID, questionID).
		Count(&incorrectCount).Error
	if err != nil {
		return false, err
	}

	// Nếu tất cả câu trả lời đều đúng, không cần làm gì
	if incorrectCount == 0 {
		return false, nil
	}

	// Nếu có câu trả lời sai, xóa tất cả câu trả lời cũ
	err = db.MasterDB.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
		Delete(nil).Error
	if err != nil {
		return false, err
	}

	// Cần lưu câu trả lời mới
	return true, nil
}

// CheckAndCleanExistingAnswersWithTx kiểm tra và xóa các câu trả lời cũ trong transaction
func CheckAndCleanExistingAnswersWithTx(tx *gorm.DB, table string, homeworkID, userID, questionID int64) (bool, error) {
	// Kiểm tra xem đã có câu trả lời chưa
	var count int64
	err := tx.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	// Nếu chưa có câu trả lời nào, cần lưu mới
	if count == 0 {
		return true, nil
	}

	// Kiểm tra xem có câu trả lời sai không
	var incorrectCount int64
	err = tx.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false", homeworkID, userID, questionID).
		Count(&incorrectCount).Error
	if err != nil {
		return false, err
	}

	// Nếu tất cả câu trả lời đều đúng, không cần làm gì
	if incorrectCount == 0 {
		return false, nil
	}

	// Nếu có câu trả lời sai, xóa tất cả câu trả lời cũ
	err = tx.Table(table).
		Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
		Delete(nil).Error
	if err != nil {
		return false, err
	}

	// Cần lưu câu trả lời mới
	return true, nil
}
