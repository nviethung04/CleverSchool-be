package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"
)

type HomeworkUserSkipQuestionRepository interface {
	Create(skipQuestion *models.HomeworkUserSkipQuestion) error
	GetByHomeworkUserQuestion(homeworkID, userID, questionID int64) (*models.HomeworkUserSkipQuestion, error)
	Update(skipQuestion *models.HomeworkUserSkipQuestion) error
	ListQuestionIDs(homeworkID, userID int64) ([]int64, error)
	UpdateDidItAgain(homeworkID, userID, questionID int64) error
	ListSkippedButRedoneQuestionIDs(homeworkID, userID int64) ([]int64, error)
	ListSkippedNotRedoneQuestionIDs(homeworkID, userID int64) ([]int64, error)
}

type homeworkUserSkipQuestionRepository struct{}

func NewHomeworkUserSkipQuestionRepository() HomeworkUserSkipQuestionRepository {
	return &homeworkUserSkipQuestionRepository{}
}

func (r *homeworkUserSkipQuestionRepository) Create(skipQuestion *models.HomeworkUserSkipQuestion) error {
	return db.MasterDB.Create(skipQuestion).Error
}

func (r *homeworkUserSkipQuestionRepository) GetByHomeworkUserQuestion(homeworkID, userID, questionID int64) (*models.HomeworkUserSkipQuestion, error) {
	var skipQuestion models.HomeworkUserSkipQuestion
	err := db.ReplicaDB.Where("homework_id = ? AND user_id = ? AND question_id = ?",
		homeworkID, userID, questionID).First(&skipQuestion).Error
	if err != nil {
		return nil, err
	}
	return &skipQuestion, nil
}

func (r *homeworkUserSkipQuestionRepository) Update(skipQuestion *models.HomeworkUserSkipQuestion) error {
	return db.MasterDB.Save(skipQuestion).Error
}

func (r *homeworkUserSkipQuestionRepository) ListQuestionIDs(homeworkID, userID int64) ([]int64, error) {
	var ids []int64
	err := db.MasterDB.Table("homework_user_skip_questions").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Pluck("question_id", &ids).Error
	return ids, err
}

func (r *homeworkUserSkipQuestionRepository) UpdateDidItAgain(homeworkID, userID, questionID int64) error {
	now := time.Now()
	return db.MasterDB.Model(&models.HomeworkUserSkipQuestion{}).
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND did_it_again = ?", 
			homeworkID, userID, questionID, false).
		Updates(map[string]interface{}{
			"did_it_again":     true,
			"did_it_again_at":  now,
			"updated_at":       now,
		}).Error
}

func (r *homeworkUserSkipQuestionRepository) ListSkippedButRedoneQuestionIDs(homeworkID, userID int64) ([]int64, error) {
	var ids []int64
	err := db.MasterDB.Table("homework_user_skip_questions").
		Where("homework_id = ? AND user_id = ? AND did_it_again = ?", homeworkID, userID, true).
		Pluck("question_id", &ids).Error
	return ids, err
}

func (r *homeworkUserSkipQuestionRepository) ListSkippedNotRedoneQuestionIDs(homeworkID, userID int64) ([]int64, error) {
	var ids []int64
	err := db.MasterDB.Table("homework_user_skip_questions").
		Where("homework_id = ? AND user_id = ? AND did_it_again = ?", homeworkID, userID, false).
		Pluck("question_id", &ids).Error
	return ids, err
}
