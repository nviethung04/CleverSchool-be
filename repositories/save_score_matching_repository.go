package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"gorm.io/gorm"
	"time"
)

type SaveScoreMatchingRepository interface {
	GetAnswerByID(answerID int64) (*models.AnswerMatching, error)
	SaveBatchExamQuestionUserMatching(records []*models.ExamQuestionUserMatching, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserMatching(records []*models.HomeworkQuestionUserMatching, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserMatching(records []*models.ExerciseQuestionUserMatching, tx *gorm.DB) error
}

type saveScoreMatchingRepository struct{}

func NewSaveScoreMatchingRepository() SaveScoreMatchingRepository {
	return &saveScoreMatchingRepository{}
}




func (r *saveScoreMatchingRepository) GetAnswerByID(answerID int64) (*models.AnswerMatching, error) {
	var answer models.AnswerMatching
	err := db.ReplicaDB.
		Where("id = ?", answerID).
		First(&answer).Error
	return &answer, err
}

func (r *saveScoreMatchingRepository) SaveBatchExamQuestionUserMatching(records []*models.ExamQuestionUserMatching, tx *gorm.DB) error {
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

func (r *saveScoreMatchingRepository) SaveBatchHomeworkQuestionUserMatching(records []*models.HomeworkQuestionUserMatching, tx *gorm.DB) error {
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
		Delete(&models.HomeworkQuestionUserMatching{}).Error
	if err != nil {
		return err
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreMatchingRepository) SaveBatchExerciseQuestionUserMatching(records []*models.ExerciseQuestionUserMatching, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	first := records[0]
	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserMatching{}).Error; err != nil { return err }
	return tx.CreateInBatches(records, len(records)).Error
}
