package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"gorm.io/gorm"
	"time"
)

type SaveScorePositionRepository interface {
	SaveBatchExamQuestionUserPositions(records []*models.ExamQuestionUserPosition, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserPositions(records []*models.HomeworkQuestionUserPosition, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserPositions(records []*models.ExerciseQuestionUserPosition, tx *gorm.DB) error
	FindTrueAnswerPosition(questionID int64, sortPosition int) (*models.AnswerPosition, error)
}

type saveScorePositionRepository struct{}

func NewSaveScorePositionRepository() SaveScorePositionRepository {
	return &saveScorePositionRepository{}
}




func (r *saveScorePositionRepository) FindTrueAnswerPosition(questionID int64, sortPosition int) (*models.AnswerPosition, error) {
	var answer models.AnswerPosition
	err := db.ReplicaDB.
		Where("question_id = ? AND correct_position = ?", questionID, sortPosition).
		First(&answer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Không tìm thấy cũng không phải lỗi nghiêm trọng
		}
		return nil, err
	}
	return &answer, nil
}

func (r *saveScorePositionRepository) SaveBatchExamQuestionUserPositions(records []*models.ExamQuestionUserPosition, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		rec.CreatedAt = time.Now().UTC()
	}
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(&records).Error
}

func (r *saveScorePositionRepository) SaveBatchHomeworkQuestionUserPositions(records []*models.HomeworkQuestionUserPosition, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		rec.CreatedAt = time.Now().UTC()
	}
	if tx == nil {
		tx = db.MasterDB
	}
	first := records[0]
	err := tx.Where("homework_id = ? AND user_id = ? AND question_id = ?", first.HomeworkID, first.UserID, first.QuestionID).
		Delete(&models.HomeworkQuestionUserPosition{}).Error
	if err != nil {
		return err
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScorePositionRepository) SaveBatchExerciseQuestionUserPositions(records []*models.ExerciseQuestionUserPosition, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	first := records[0]
	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserPosition{}).Error; err != nil { return err }
	return tx.CreateInBatches(records, len(records)).Error
}
