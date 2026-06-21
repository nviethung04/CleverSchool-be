package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"gorm.io/gorm"
)

func DeleteAllExamAnswersByExamIDAndUserID(examID, userID int64, tx *gorm.DB) error {
	if tx == nil {
		tx = db.MasterDB
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUser{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserFillInBlank{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserGroup{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserMatching{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserPosition{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserLabeling{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exam_id = ? AND user_id = ?", examID, userID).Delete(&models.ExamQuestionUserManualScoring{}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteAllExerciseAnswersByExerciseIDAndUserID(exerciseID, userID int64, tx *gorm.DB) error {
	if tx == nil {
		tx = db.MasterDB
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUser{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserFillInBlank{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserGroup{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserMatching{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserPosition{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserLabeling{}).Error; err != nil {
		return err
	}
	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Delete(&models.ExerciseQuestionUserManualScoring{}).Error; err != nil {
		return err
	}
	return nil
}

// ResetExerciseAttempt xóa kết quả và đáp án của HS để cho phép làm lại bài luyện tập.
// Luôn xóa theo exercise_id + user_id (không lọc lesson_id) vì completion map
// và dữ liệu cũ có thể lưu lesson_id=0 hoặc khác bài học hiện tại.
func ResetExerciseAttempt(exerciseID, userID, lessonID int64) error {
	_ = lessonID
	tx := db.MasterDB.Begin()
	if err := DeleteAllExerciseAnswersByExerciseIDAndUserID(exerciseID, userID, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Delete(&models.ExerciseUser{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("exercise_id = ? AND student_id = ?", exerciseID, userID).
		Delete(&models.ExerciseComment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
