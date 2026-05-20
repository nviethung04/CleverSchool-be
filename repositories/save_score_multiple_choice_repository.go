package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"gorm.io/gorm"
	"time"
)

type SaveScoreMultipleChoiceRepository interface {
	GetCorrectAnswersFromExamMultipleChoice(examID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error)
	GetCorrectAnswersFromLevelTestMultipleChoice(levelTestID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error)
	GetCorrectAnswersFromHomeworkMultipleChoice(homeworkID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error)
	GetCorrectAnswersFromExerciseMultipleChoice(exerciseID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error)
	GetAnswersByIDsMultipleChoice(answerIDs []int64, questionID int64) ([]*models.Answer, error)

	SaveExamQuestionUserMultipleChoice(record *models.ExamQuestionUser, tx *gorm.DB) error
	SaveHomeworkQuestionUserMultipleChoice(record *models.HomeworkQuestionUser, tx *gorm.DB) error
	SaveLevelTestQuestionUserMultipleChoice(record *models.LevelTestQuestionUser, tx *gorm.DB) error
	SaveBatchExamQuestionUserMultipleChoice(records []*models.ExamQuestionUser, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserMultipleChoice(records []*models.HomeworkQuestionUser, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserMultipleChoice(records []*models.ExerciseQuestionUser, tx *gorm.DB) error
	SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error
	CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error)
}

type saveScoreMultipleChoiceRepository struct{}

func NewSaveScoreMultipleChoiceRepository() SaveScoreMultipleChoiceRepository {
	return &saveScoreMultipleChoiceRepository{}
}

// repository/save_score_repository.go

func (r *saveScoreMultipleChoiceRepository) GetCorrectAnswersFromExamMultipleChoice(examID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error) {
	var results []*dto.AnswerWithScoreMultipleChoice
	err := db.ReplicaDB.
		Table("answers").
		Select("answers.*, 1.0 AS score").
		Where("answers.question_id = ? AND answers.is_correct = true", questionID).
		Scan(&results).Error

	return results, err
}

func (r *saveScoreMultipleChoiceRepository) GetCorrectAnswersFromLevelTestMultipleChoice(levelTestID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error) {
	var results []*dto.AnswerWithScoreMultipleChoice
	err := db.ReplicaDB.
		Table("answers").
		Select("answers.*, level_test_questions.score AS score").
		Joins("JOIN level_test_questions ON level_test_questions.question_id = answers.question_id").
		Where("level_test_questions.level_test_id = ? AND answers.question_id = ? AND answers.is_correct = true", levelTestID, questionID).
		Scan(&results).Error

	return results, err
}

func (r *saveScoreMultipleChoiceRepository) GetCorrectAnswersFromHomeworkMultipleChoice(homeworkID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error) {
	var results []*dto.AnswerWithScoreMultipleChoice
	err := db.ReplicaDB.
		Table("answers").
		Select("answers.*, 1.0 AS score").
		Where("answers.question_id = ? AND answers.is_correct = true", questionID).
		Scan(&results).Error

	return results, err
}

func (r *saveScoreMultipleChoiceRepository) GetCorrectAnswersFromExerciseMultipleChoice(exerciseID, questionID int64) ([]*dto.AnswerWithScoreMultipleChoice, error) {
	var results []*dto.AnswerWithScoreMultipleChoice
	err := db.ReplicaDB.
		Table("answers").
		Select("answers.*, 1.0 AS score").
		Where("answers.question_id = ? AND answers.is_correct = true", questionID).
		Scan(&results).Error
	return results, err
}

func (r *saveScoreMultipleChoiceRepository) GetAnswersByIDsMultipleChoice(answerIDs []int64, questionID int64) ([]*models.Answer, error) {
	var answers []*models.Answer
	err := db.ReplicaDB.Where("id IN ? AND question_id = ?", answerIDs, questionID).Find(&answers).Error
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *saveScoreMultipleChoiceRepository) SaveExamQuestionUserMultipleChoice(record *models.ExamQuestionUser, tx *gorm.DB) error {
	record.CreatedAt = time.Now().UTC()
	if tx == nil {
		tx = db.MasterDB
	}

	err := tx.Where("exam_id = ? AND user_id = ?", record.ExamID, record.UserID).
		Delete(&models.ExamQuestionUser{}).Error
	if err != nil {
		return err
	}

	return tx.Create(record).Error
}

func (r *saveScoreMultipleChoiceRepository) SaveHomeworkQuestionUserMultipleChoice(record *models.HomeworkQuestionUser, tx *gorm.DB) error {
	record.CreatedAt = time.Now().UTC()
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(record).Error
}

func (r *saveScoreMultipleChoiceRepository) SaveLevelTestQuestionUserMultipleChoice(record *models.LevelTestQuestionUser, tx *gorm.DB) error {
	record.CreatedAt = time.Now().UTC()
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(record).Error
}

func (r *saveScoreMultipleChoiceRepository) SaveBatchExamQuestionUserMultipleChoice(records []*models.ExamQuestionUser, tx *gorm.DB) error {
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

func (r *saveScoreMultipleChoiceRepository) SaveBatchHomeworkQuestionUserMultipleChoice(records []*models.HomeworkQuestionUser, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		rec.CreatedAt = time.Now().UTC()
	}
	if tx == nil {
		tx = db.MasterDB
	}
	// Xóa đáp án cũ
	err := tx.Where("homework_id = ? AND user_id = ? AND question_id = ?", records[0].HomeworkID, records[0].UserID, records[0].QuestionID).
		Delete(&models.HomeworkQuestionUser{}).Error
	if err != nil {
		return err
	}
	return tx.Create(&records).Error
}

func (r *saveScoreMultipleChoiceRepository) SaveBatchExerciseQuestionUserMultipleChoice(records []*models.ExerciseQuestionUser, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		rec.CreatedAt = time.Now().UTC()
	}
	if tx == nil {
		tx = db.MasterDB
	}
	// Xóa đáp án cũ
	err := tx.Where("exercise_id = ? AND user_id = ?", records[0].ExerciseID, records[0].UserID).
		Delete(&models.ExerciseQuestionUser{}).Error
	if err != nil {
		return err
	}
	return tx.Create(&records).Error
}

func (r *saveScoreMultipleChoiceRepository) SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error {
	record.CreatedAt = time.Now().UTC()
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(record).Error
}

func (r *saveScoreMultipleChoiceRepository) CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error) {
	var count int64
	err := db.MasterDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ?", homeworkID, userID, questionID, lessonID).
		Count(&count).Error
	return count, err
}
