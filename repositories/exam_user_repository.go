package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ExamUserRepository interface {
	SaveExamUser(examUser *models.ExamUser) error
	SumScoreFillInBlank(examID, userID int64) (float64, error)
	SumScoreGroup(examID, userID int64) (float64, error)
	SumScoreLabeling(examID, userID int64) (float64, error)
	SumScoreManual(examID, userID int64) (float64, error)
	SumScoreMatching(examID, userID int64) (float64, error)
	SumScorePosition(examID, userID int64) (float64, error)
	SumScoreUser(examID, userID int64) (float64, error)
	UpdateExamUserScore(examID, userID int64, score float64) error
	GetCorrectCountByTable(table string, examID, userID int64) (map[int64]int, error)
	GetManualScoringByExam(examID, userID int64) (map[int64]float64, error)
	GetAnswerCountByTable(table string, examID, userID int64) (map[int64]int, error)
	UpdateExamUserRatio(examID, userID int64, ratio float64) error
	// Exercise methods
	SaveExerciseUser(exerciseUser *models.ExerciseUser) error
	SumScoreFillInBlankExercise(exerciseID, userID int64) (float64, error)
	SumScoreGroupExercise(exerciseID, userID int64) (float64, error)
	SumScoreLabelingExercise(exerciseID, userID int64) (float64, error)
	SumScoreManualExercise(exerciseID, userID int64) (float64, error)
	SumScoreMatchingExercise(exerciseID, userID int64) (float64, error)
	SumScorePositionExercise(exerciseID, userID int64) (float64, error)
	SumScoreUserExercise(exerciseID, userID int64) (float64, error)
	UpdateExerciseUserScore(exerciseID, userID int64, score float64) error
	GetCorrectCountByTableExercise(table string, exerciseID, userID int64) (map[int64]int, error)
	GetManualScoringByExercise(exerciseID, userID int64) (map[int64]float64, error)
	GetAnswerCountByTableExercise(table string, exerciseID, userID int64) (map[int64]int, error)
	UpdateExerciseUserRatio(exerciseID, userID int64, ratio float64) error
	UpdateOrCreate(examUser *models.ExamUser) error
}

type examUserRepository struct {
	db *gorm.DB
}

func NewExamUserRepository() ExamUserRepository {
	return &examUserRepository{}
}

func (r *examUserRepository) SaveExamUser(examUser *models.ExamUser) error {
	// Kiểm tra xem đã có bản ghi cho bài kiểm tra này chưa
	var existingExamUser models.ExamUser
	err := db.MasterDB.Where("exam_id = ? AND user_id = ?", examUser.ExamID, examUser.UserID).First(&existingExamUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Nếu đã có bản ghi thì cập nhật
	if existingExamUser.ID != 0 {
		return db.MasterDB.Model(&existingExamUser).Updates(map[string]interface{}{
			"score":              examUser.Score,
			"time":               examUser.Time,
			"has_manual_scoring": examUser.HasManualScoring,
			"updated_at":         time.Now().UTC(),
		}).Error
	}

	// Nếu chưa có bản ghi thì tạo mới
	now := time.Now().UTC()
	examUser.CreatedAt = now
	examUser.UpdatedAt = now
	return db.MasterDB.Create(examUser).Error
}

func (r *examUserRepository) SaveExerciseUser(exerciseUser *models.ExerciseUser) error {
	// Kiểm tra xem đã có bản ghi cho bài tập này chưa
	var existingExerciseUser models.ExerciseUser
	err := db.MasterDB.Where("exercise_id = ? AND user_id = ?", exerciseUser.ExerciseID, exerciseUser.UserID).First(&existingExerciseUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Nếu đã có bản ghi thì cập nhật
	if existingExerciseUser.ID != 0 {
		return db.MasterDB.Model(&existingExerciseUser).Updates(map[string]interface{}{
			"score":              exerciseUser.Score,
			"time":               exerciseUser.Time,
			"has_manual_scoring": exerciseUser.HasManualScoring,
			"updated_at":         time.Now().UTC(),
		}).Error
	}

	// Nếu chưa có bản ghi thì tạo mới
	now := time.Now().UTC()
	exerciseUser.CreatedAt = now
	exerciseUser.UpdatedAt = now
	return db.MasterDB.Create(exerciseUser).Error
}

func (r *examUserRepository) SumScoreFillInBlank(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_fill_in_blanks").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScoreGroup(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_groups").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScoreLabeling(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_labelings").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScoreManual(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_manual_scoring").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScoreMatching(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_matchings").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScorePosition(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_user_positions").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) SumScoreUser(examID, userID int64) (float64, error) {
	var sum float64
	err := db.ReplicaDB.Table("exam_question_users").
		Select("COALESCE(SUM(score),0)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *examUserRepository) UpdateExamUserScore(examID, userID int64, score float64) error {
	dbConn := db.MasterDB
	var examUserID int64
	err := dbConn.Table("exam_users").
		Select("id").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&examUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if examUserID > 0 {
		return dbConn.Table("exam_users").
			Where("id = ?", examUserID).
			Updates(map[string]interface{}{
				"score":      score,
				"updated_at": now,
			}).Error
	} else {
		return dbConn.Table("exam_users").
			Create(map[string]interface{}{
				"exam_id":    examID,
				"user_id":    userID,
				"score":      score,
				"created_at": now,
				"updated_at": now,
			}).Error
	}
}

//Xử lý chấm điểm theo tỉ lệ


func (r *examUserRepository) GetCorrectCountByTable(table string, examID, userID int64) (map[int64]int, error) {
	result := make(map[int64]int)
	rows, err := db.ReplicaDB.Table(table).
		Select("question_id, COUNT(*)").
		Where("exam_id = ? AND user_id = ? AND is_correct = true", examID, userID).
		Group("question_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var cnt int
		if err := rows.Scan(&qid, &cnt); err == nil {
			result[qid] = cnt
		}
	}
	return result, nil
}

func (r *examUserRepository) GetAnswerCountByTable(table string, examID, userID int64) (map[int64]int, error) {
	result := make(map[int64]int)
	rows, err := db.ReplicaDB.Table(table).
		Select("question_id, COUNT(*)").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Group("question_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var cnt int
		if err := rows.Scan(&qid, &cnt); err == nil {
			result[qid] = cnt
		}
	}
	return result, nil
}

func (r *examUserRepository) GetManualScoringByExam(examID, userID int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	rows, err := db.ReplicaDB.Table("exam_question_user_manual_scoring").
		Select("question_id, score").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var score float64
		if err := rows.Scan(&qid, &score); err == nil {
			result[qid] = score
		}
	}
	return result, nil
}

func (r *examUserRepository) UpdateExamUserRatio(examID, userID int64, ratio float64) error {
	dbConn := db.MasterDB
	var examUserID int64
	err := dbConn.Table("exam_users").
		Select("id").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&examUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if examUserID > 0 {
		return dbConn.Table("exam_users").
			Where("id = ?", examUserID).
			Updates(map[string]interface{}{
				"ratio":      ratio,
				"updated_at": now,
			}).Error
	} else {
		return dbConn.Table("exam_users").
			Create(map[string]interface{}{
				"exam_id":    examID,
				"user_id":    userID,
				"ratio":      ratio,
				"created_at": now,
				"updated_at": now,
			}).Error
	}
}

// Exercise methods - tạm thời return 0 hoặc nil để tránh lỗi compile
func (r *examUserRepository) SumScoreFillInBlankExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScoreGroupExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScoreLabelingExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScoreManualExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScoreMatchingExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScorePositionExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) SumScoreUserExercise(exerciseID, userID int64) (float64, error) {
	return 0, nil
}

func (r *examUserRepository) UpdateExerciseUserScore(exerciseID, userID int64, score float64) error {
	// Kiểm tra xem bản ghi có tồn tại không
	var count int64
	err := db.MasterDB.Model(&models.ExerciseUser{}).
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if count > 0 {
		// Bản ghi đã tồn tại, cập nhật
		return db.MasterDB.Model(&models.ExerciseUser{}).
			Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
			Updates(map[string]interface{}{
				"score":      score,
				"updated_at": now,
			}).Error
	} else {
		// Bản ghi chưa tồn tại, tạo mới
		return db.MasterDB.Create(&models.ExerciseUser{
			ExerciseID: exerciseID,
			UserID:     userID,
			Score:      &score,
			CreatedAt:  now,
			UpdatedAt:  now,
		}).Error
	}
}

func (r *examUserRepository) GetCorrectCountByTableExercise(table string, exerciseID, userID int64) (map[int64]int, error) {
    result := make(map[int64]int)
    rows, err := db.ReplicaDB.Table(table).
        Select("question_id, COUNT(*)").
        Where("exercise_id = ? AND user_id = ? AND is_correct = true", exerciseID, userID).
        Group("question_id").
        Rows()
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var qid int64
        var cnt int
        if err := rows.Scan(&qid, &cnt); err == nil {
            result[qid] = cnt
        }
    }
    return result, nil
}

func (r *examUserRepository) GetManualScoringByExercise(exerciseID, userID int64) (map[int64]float64, error) {
    result := make(map[int64]float64)
    rows, err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").
        Select("question_id, score").
        Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
        Rows()
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var qid int64
        var score float64
        if err := rows.Scan(&qid, &score); err == nil {
            result[qid] = score
        }
    }
    return result, nil
}

func (r *examUserRepository) GetAnswerCountByTableExercise(table string, exerciseID, userID int64) (map[int64]int, error) {
    result := make(map[int64]int)
    rows, err := db.ReplicaDB.Table(table).
        Select("question_id, COUNT(*)").
        Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
        Group("question_id").
        Rows()
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var qid int64
        var cnt int
        if err := rows.Scan(&qid, &cnt); err == nil {
            result[qid] = cnt
        }
    }
    return result, nil
}

func (r *examUserRepository) UpdateExerciseUserRatio(exerciseID, userID int64, ratio float64) error {
	// Kiểm tra xem bản ghi có tồn tại không
	var count int64
	err := db.MasterDB.Model(&models.ExerciseUser{}).
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if count > 0 {
		// Bản ghi đã tồn tại, cập nhật
		return db.MasterDB.Model(&models.ExerciseUser{}).
			Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
			Updates(map[string]interface{}{
				"ratio":      ratio,
				"updated_at": now,
			}).Error
	} else {
		// Bản ghi chưa tồn tại, tạo mới
		return db.MasterDB.Create(&models.ExerciseUser{
			ExerciseID: exerciseID,
			UserID:     userID,
			Ratio:      &ratio,
			CreatedAt:  now,
			UpdatedAt:  now,
		}).Error
	}
}

func (r *examUserRepository) UpdateOrCreate(examUser *models.ExamUser) error {
	if examUser == nil {
		return fmt.Errorf("examUser cannot be nil")
	}

	var existing models.ExamUser
	err := db.ReplicaDB.
		Where("exam_id = ? AND lesson_id = ? AND user_id = ?",
			examUser.ExamID, examUser.LessonID, examUser.UserID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Chưa có => tạo mới
			if err := db.MasterDB.Create(examUser).Error; err != nil {
				return fmt.Errorf("failed to create exam user: %w", err)
			}
			return nil
		}
		// Lỗi khác
		return fmt.Errorf("failed to query exam user: %w", err)
	}

	// Có rồi => update FileInfos
	err = db.ReplicaDB.Model(&existing).
		Updates(map[string]interface{}{
			"file_infos":  examUser.FileInfos,
			"has_manual_scoring":  examUser.HasManualScoring,
			"score":       examUser.Score,
			"ratio":       examUser.Ratio,
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update exam user: %w", err)
	}

	return nil
}
