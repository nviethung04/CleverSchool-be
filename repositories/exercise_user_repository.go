package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ExerciseUserRepository interface {
	SaveExerciseUser(exerciseUser *models.ExerciseUser) error
	SumScoreFillInBlank(exerciseID, userID int64) (float64, error)
	SumScoreGroup(exerciseID, userID int64) (float64, error)
	SumScoreLabeling(exerciseID, userID int64) (float64, error)
	SumScoreManual(exerciseID, userID int64) (float64, error)
	SumScoreMatching(exerciseID, userID int64) (float64, error)
	SumScorePosition(exerciseID, userID int64) (float64, error)
	SumScoreUser(exerciseID, userID int64) (float64, error)
	UpdateExerciseUserScore(exerciseID, userID int64, score float64) error
	GetCorrectCountByTable(table string, exerciseID, userID int64) (map[int64]int, error)
	GetManualScoringByExercise(exerciseID, userID int64) (map[int64]float64, error)
	GetAnswerCountByTable(table string, exerciseID, userID int64) (map[int64]int, error)
	UpdateExerciseUserRatio(exerciseID, userID int64, ratio float64) error
	UpdateOrCreate(exerciseUser *models.ExerciseUser) error
}

type exerciseUserRepository struct {
	db *gorm.DB
}

func NewExerciseUserRepository() ExerciseUserRepository {
	return &exerciseUserRepository{}
}

func (r *exerciseUserRepository) SaveExerciseUser(exerciseUser *models.ExerciseUser) error {
	return db.MasterDB.Save(exerciseUser).Error
}

func (r *exerciseUserRepository) SumScoreFillInBlank(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_fill_in_blanks").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScoreGroup(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_groups").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScoreLabeling(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_labelings").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScoreManual(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").
		Where("exercise_id = ? AND user_id = ? AND is_scored = true", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScoreMatching(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_matchings").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScorePosition(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_user_positions").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) SumScoreUser(exerciseID, userID int64) (float64, error) {
	var total float64
	err := db.ReplicaDB.Table("exercise_question_users").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Select("COALESCE(SUM(score),0)").
		Scan(&total).Error
	return total, err
}

func (r *exerciseUserRepository) UpdateExerciseUserScore(exerciseID, userID int64, score float64) error {
	dbConn := db.MasterDB
	var id int64
	err := dbConn.Table("exercise_users").Select("id").Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Scan(&id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if id > 0 {
		return dbConn.Table("exercise_users").Where("id = ?", id).Updates(map[string]interface{}{
			"score":      score,
			"updated_at": now,
		}).Error
	}
	return dbConn.Table("exercise_users").Create(map[string]interface{}{
		"exercise_id": exerciseID,
		"user_id":     userID,
		"score":       score,
		"created_at":  now,
		"updated_at":  now,
	}).Error
}

func (r *exerciseUserRepository) GetCorrectCountByTable(table string, exerciseID, userID int64) (map[int64]int, error) {
	rows := make([]struct{
		QuestionID int64
		Count int
	}, 0)
	err := db.ReplicaDB.Table(table).
		Select("question_id, COUNT(*) as count").
		Where("exercise_id = ? AND user_id = ? AND is_correct = true", exerciseID, userID).
		Group("question_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int)
	for _, r := range rows {
		result[r.QuestionID] = r.Count
	}
	return result, nil
}

func (r *exerciseUserRepository) GetAnswerCountByTable(table string, exerciseID, userID int64) (map[int64]int, error) {
	rows := make([]struct{
		QuestionID int64
		Count int
	}, 0)
	err := db.ReplicaDB.Table(table).
		Select("question_id, COUNT(*) as count").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Group("question_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int)
	for _, r := range rows {
		result[r.QuestionID] = r.Count
	}
	return result, nil
}

func (r *exerciseUserRepository) GetManualScoringByExercise(exerciseID, userID int64) (map[int64]float64, error) {
	rows := make([]struct{
		QuestionID int64
		Score float64
	}, 0)
	err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").
		Select("question_id, score").
		Where("exercise_id = ? AND user_id = ? AND is_scored = true", exerciseID, userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]float64)
	for _, r := range rows {
		result[r.QuestionID] = r.Score
	}
	return result, nil
}

func (r *exerciseUserRepository) UpdateExerciseUserRatio(exerciseID, userID int64, ratio float64) error {
	dbConn := db.MasterDB
	var id int64
	err := dbConn.Table("exercise_users").
		Select("id").
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Scan(&id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if id > 0 {
		return dbConn.Table("exercise_users").
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"ratio":      ratio,
				"updated_at": now,
			}).Error
	}
	return dbConn.Table("exercise_users").
		Create(map[string]interface{}{
			"exercise_id": exerciseID,
			"user_id":     userID,
			"ratio":       ratio,
			"created_at":  now,
			"updated_at":  now,
		}).Error
}

func (r *exerciseUserRepository) UpdateOrCreate(exerciseUser *models.ExerciseUser) error {
	if exerciseUser == nil {
		return fmt.Errorf("exerciseUser cannot be nil")
	}

	var existing models.ExerciseUser
	err := db.MasterDB.
		Where("exercise_id = ? AND lesson_id = ? AND user_id = ?",
			exerciseUser.ExerciseID, exerciseUser.LessonID, exerciseUser.UserID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Chưa có => tạo mới
			if err := db.ReplicaDB.Create(exerciseUser).Error; err != nil {
				return fmt.Errorf("failed to create exercise user: %w", err)
			}
			return nil
		}
		// Lỗi khác
		return fmt.Errorf("failed to query exercise user: %w", err)
	}

	// Có rồi => update FileInfos
	err = db.ReplicaDB.Model(&existing).
		Updates(map[string]interface{}{
			"file_infos":  exerciseUser.FileInfos,
			"has_manual_scoring":  exerciseUser.HasManualScoring,
			"score":       exerciseUser.Score,
			"ratio":       exerciseUser.Ratio,
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update exercise user: %w", err)
	}

	return nil
}
