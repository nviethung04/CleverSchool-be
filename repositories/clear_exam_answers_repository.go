package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
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
