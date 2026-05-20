package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"gorm.io/gorm"
	"time"
)

type SaveScoreLabelingRepository interface {
	GetExamQuestionScoreLabeling(examID, questionID int64) (float64, error)
	GetHomeworkQuestionScoreLabeling(homeworkID, questionID int64) (float64, error)
	GetExerciseQuestionScoreLabeling(exerciseID, questionID int64) (float64, error)
	// GetAnswerByID GetLevelTestQuestionScoreLabeling(levelTestID, questionID int64) (float64, error)
	GetAnswerByID(id int64) (*models.AnswerCoordinates, error)
	SaveBatchExamQuestionUserLabeling(records []*models.ExamQuestionUserLabeling, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserLabeling(records []*models.HomeworkQuestionUserLabeling, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserLabeling(records []*models.ExerciseQuestionUserLabeling, tx *gorm.DB) error
	//SaveBatchLevelTestQuestionUserLabeling(records []*models.LevelTestQuestionUserLabeling, tx *gorm.DB) error
}

type saveScoreLabelingRepository struct{}

func NewSaveScoreLabelingRepository() SaveScoreLabelingRepository {
	return &saveScoreLabelingRepository{}
}

func (r *saveScoreLabelingRepository) GetExamQuestionScoreLabeling(examID, questionID int64) (float64, error) {
	var examQuestion models.ExamQuestion
	err := db.MasterDB.Where("exam_id = ? AND question_id = ?", examID, questionID).First(&examQuestion).Error
	if err != nil {
		return 0, err
	}
	return examQuestion.Score, nil
}

func (r *saveScoreLabelingRepository) GetHomeworkQuestionScoreLabeling(homeworkID, questionID int64) (float64, error) {
	var homeworkQuestion models.HomeworkQuestion
	err := db.MasterDB.Where("homework_id = ? AND question_id = ?", homeworkID, questionID).First(&homeworkQuestion).Error
	if err != nil {
		return 0, err
	}
	return homeworkQuestion.Score, nil
}

func (r *saveScoreLabelingRepository) GetExerciseQuestionScoreLabeling(exerciseID, questionID int64) (float64, error) {
	var exerciseQuestion models.ExerciseQuestion
	err := db.MasterDB.Where("exercise_id = ? AND question_id = ?", exerciseID, questionID).First(&exerciseQuestion).Error
	if err != nil {
		return 0, err
	}
	return exerciseQuestion.Score, nil
}

//func (r *saveScoreLabelingRepository) GetLevelTestQuestionScoreLabeling(levelTestID, questionID int64) (float64, error) {
//	var levelTestQuestion models.LevelTestQuestion
//	err := db.MasterDB.Where("level_test_id = ? AND question_id = ?", levelTestID, questionID).First(&levelTestQuestion).Error
//	if err != nil {
//		return 0, err
//	}
//	return levelTestQuestion.Score, nil
//}

func (r *saveScoreLabelingRepository) GetAnswerByID(id int64) (*models.AnswerCoordinates, error) {
	var answer models.AnswerCoordinates
	err := db.MasterDB.Where("id = ?", id).First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (r *saveScoreLabelingRepository) SaveBatchExamQuestionUserLabeling(records []*models.ExamQuestionUserLabeling, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		rec.CreatedAt = time.Now().UTC()
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreLabelingRepository) SaveBatchHomeworkQuestionUserLabeling(records []*models.HomeworkQuestionUserLabeling, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, rec := range records {
		rec.CreatedAt = now
	}
	first := records[0]
	if tx == nil {
		tx = db.MasterDB
	}
	err := tx.Where("homework_id = ? AND user_id = ? AND question_id = ?", first.HomeworkID, first.UserID, first.QuestionID).
		Delete(&models.HomeworkQuestionUserLabeling{}).Error
	if err != nil {
		return err
	}
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreLabelingRepository) SaveBatchExerciseQuestionUserLabeling(records []*models.ExerciseQuestionUserLabeling, tx *gorm.DB) error {
	if len(records) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, rec := range records {
		rec.CreatedAt = now
	}
	first := records[0]
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserLabeling{}).Error; err != nil {
		return err
	}

	if err := tx.CreateInBatches(records, len(records)).Error; err != nil {
		return err
	}
	return nil
}

//func (r *saveScoreLabelingRepository) SaveBatchLevelTestQuestionUserLabeling(records []*models.LevelTestQuestionUserLabeling, tx *gorm.DB) error {
//	if len(records) == 0 {
//		return nil
//	}
//	return tx.CreateInBatches(records, len(records)).Error
//}
