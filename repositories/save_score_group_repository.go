package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"gorm.io/gorm"
	"time"
)

type SaveScoreGroupRepository interface {
	GetExamQuestionScoreGroup(examID, questionID int64) (float64, error)
	GetHomeworkQuestionScoreGroup(homeworkID, questionID int64) (float64, error)
	GetExerciseQuestionScoreGroup(exerciseID, questionID int64) (float64, error)
	GetAnswerByID(id int64) (*models.AnswerGroup, error)
	GetGroupByID(id int64) (*models.GroupAnswer, error)
	GetAnswerGroupByAnswerID(answerID int64) (*models.AnswerGroup, error)
	SaveBatchExamQuestionUserGroup(records []*models.ExamQuestionUserGroup, tx *gorm.DB) error
	SaveBatchHomeworkQuestionUserGroup(records []*models.HomeworkQuestionUserGroup, tx *gorm.DB) error
	SaveBatchExerciseQuestionUserGroup(records []*models.ExerciseQuestionUserGroup, tx *gorm.DB) error
}

type saveScoreGroupRepository struct{}

func NewSaveScoreGroupRepository() SaveScoreGroupRepository {
	return &saveScoreGroupRepository{}
}

func (r *saveScoreGroupRepository) GetExamQuestionScoreGroup(examID, questionID int64) (float64, error) {
	var examQuestion models.ExamQuestion
	err := db.MasterDB.Where("exam_id = ? AND question_id = ?", examID, questionID).First(&examQuestion).Error
	if err != nil {
		return 0, err
	}
	return examQuestion.Score, nil
}

func (r *saveScoreGroupRepository) GetHomeworkQuestionScoreGroup(homeworkID, questionID int64) (float64, error) {
	var homeworkQuestion models.HomeworkQuestion
	err := db.MasterDB.Where("homework_id = ? AND question_id = ?", homeworkID, questionID).First(&homeworkQuestion).Error
	if err != nil {
		return 0, err
	}
	return homeworkQuestion.Score, nil
}

func (r *saveScoreGroupRepository) GetExerciseQuestionScoreGroup(exerciseID, questionID int64) (float64, error) {
	var exerciseQuestion models.ExerciseQuestion
	err := db.MasterDB.Where("exercise_id = ? AND question_id = ?", exerciseID, questionID).First(&exerciseQuestion).Error
	if err != nil {
		return 0, err
	}
	return exerciseQuestion.Score, nil
}

func (r *saveScoreGroupRepository) GetAnswerByID(id int64) (*models.AnswerGroup, error) {
	var answer models.AnswerGroup
	err := db.MasterDB.Where("id = ?", id).First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func (r *saveScoreGroupRepository) GetGroupByID(id int64) (*models.GroupAnswer, error) {
	var group models.GroupAnswer
	err := db.MasterDB.Where("id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *saveScoreGroupRepository) GetAnswerGroupByAnswerID(answerID int64) (*models.AnswerGroup, error) {
	var answerGroup models.AnswerGroup
	err := db.MasterDB.Where("id = ?", answerID).First(&answerGroup).Error
	if err != nil {
		return nil, err
	}
	return &answerGroup, nil
}

func (r *saveScoreGroupRepository) SaveBatchExamQuestionUserGroup(records []*models.ExamQuestionUserGroup, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreGroupRepository) SaveBatchHomeworkQuestionUserGroup(records []*models.HomeworkQuestionUserGroup, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	first := records[0]
	if err := tx.Where("homework_id = ? AND user_id = ? AND question_id = ?", first.HomeworkID, first.UserID, first.QuestionID).
		Delete(&models.HomeworkQuestionUserGroup{}).Error; err != nil { return err }
	return tx.CreateInBatches(records, len(records)).Error
}

func (r *saveScoreGroupRepository) SaveBatchExerciseQuestionUserGroup(records []*models.ExerciseQuestionUserGroup, tx *gorm.DB) error {
	if len(records) == 0 { return nil }
	for _, rec := range records { rec.CreatedAt = time.Now().UTC() }
	if tx == nil { tx = db.MasterDB }
	first := records[0]
	if err := tx.Where("exercise_id = ? AND user_id = ? AND question_id = ?", first.ExerciseID, first.UserID, first.QuestionID).
		Delete(&models.ExerciseQuestionUserGroup{}).Error; err != nil { return err }
	return tx.CreateInBatches(records, len(records)).Error
}
