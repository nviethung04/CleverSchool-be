package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"errors"
	"gorm.io/gorm"
	"time"
)

type SaveScoreFillInBlankRepository interface {
	SaveBatchExamQuestionUserFillInBlanks(records []*models.ExamQuestionUserFillInBlank, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserFillInBlanks(records []*models.HomeworkQuestionUserFillInBlank, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserFillInBlanks(records []*models.ExerciseQuestionUserFillInBlank, tx *gorm.DB) error
	FindTrueAnswerFillInBlank(questionID int64, sortPosition int) (*models.AnswerPosition, error)
	SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error
	CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error)
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
	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserFillInBlank{}).Error; err != nil {
		return err
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreFillInBlankRepository) SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error {
	record.CreatedAt = time.Now().UTC()
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(record).Error
}

func (r *saveScoreFillInBlankRepository) CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error) {
	var count int64
	err := db.MasterDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ?", homeworkID, userID, questionID, lessonID).
		Count(&count).Error
	return count, err
}
