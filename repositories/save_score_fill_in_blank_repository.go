package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"gorm.io/gorm"
	"time"
)

type SaveScoreFillInBlankRepository interface {
	SaveBatchExamQuestionUserFillInBlanks(records []*models.ExamQuestionUserFillInBlank, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserFillInBlanks(records []*models.HomeworkQuestionUserFillInBlank, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserFillInBlanks(records []*models.ExerciseQuestionUserFillInBlank, tx *gorm.DB) error
	FindTrueAnswerFillInBlank(questionID int64, sortPosition int) (*models.AnswerPosition, error)
}

type saveScoreFillInBlankRepository struct{}

func NewSaveScoreFillInBlankRepository() SaveScoreFillInBlankRepository {
	return &saveScoreFillInBlankRepository{}
}




func (r *saveScoreFillInBlankRepository) FindTrueAnswerFillInBlank(questionID int64, sortPosition int) (*models.AnswerPosition, error) {
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

func (r *saveScoreFillInBlankRepository) SaveBatchExamQuestionUserFillInBlanks(records []*models.ExamQuestionUserFillInBlank, tx *gorm.DB) error {
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

func (r *saveScoreFillInBlankRepository) SaveBatchHomeworkQuestionUserFillInBlanks(records []*models.HomeworkQuestionUserFillInBlank, tx *gorm.DB) error {
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
		Delete(&models.HomeworkQuestionUserFillInBlank{}).Error
	if err != nil {
		return err
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreFillInBlankRepository) SaveBatchExerciseQuestionUserFillInBlanks(records []*models.ExerciseQuestionUserFillInBlank, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	first := records[0]
	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserFillInBlank{}).Error; err != nil { return err }
	return tx.CreateInBatches(records, len(records)).Error
}
