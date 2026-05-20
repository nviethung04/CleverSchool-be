package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"time"
)

type ManualScoringRepository interface {
	SaveExamManualScoringRepository(scoring *models.ExamQuestionUserManualScoring) error
	SaveExerciseManualScoringRepository(scoring *models.ExerciseQuestionUserManualScoring) error
	SaveHomeworkManualScoringRepository(scoring *models.HomeworkQuestionUserManualScoring) error
}

type SaveScoreManualScoringRepository interface {
	UpdateManualScoring(examID, userID, questionID int64, score float64, scoringBy int64) error
	UpdateManualScoringExercise(exerciseID, userID, questionID int64, score float64, scoringBy int64) error
    UpdateManualScoringHomework(homeworkID, userID, questionID int64, score float64, scoringBy int64) error
	CheckUnscoredQuestions(examID, userID int64) (bool, error)
	CheckUnscoredQuestionsExercise(exerciseID, userID int64) (bool, error)
    CheckUnscoredQuestionsHomework(homeworkID, userID int64) (bool, error)
	UpdateExamUserManualScoringStatus(examID, userID int64, hasManualScoring bool) error
	UpdateExerciseUserManualScoringStatus(exerciseID, userID int64, hasManualScoring bool) error
    UpdateHomeworkUserManualScoringStatus(homeworkID, userID int64, hasManualScoring bool) error
}

type manualScoringRepository struct {
}

type saveScoreManualScoringRepository struct {
}

func NewManualScoringRepository() ManualScoringRepository {
	return &manualScoringRepository{}
}

func NewSaveScoreManualScoringRepository() SaveScoreManualScoringRepository {
	return &saveScoreManualScoringRepository{}
}

// SaveExamManualScoringRepository SaveExamManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài thi
func (r *manualScoringRepository) SaveExamManualScoringRepository(scoring *models.ExamQuestionUserManualScoring) error {
	if scoring != nil {
		t := time.Now().UTC()
		scoring.CreatedAt = &t
	}
	err := db.MasterDB.Where("exam_id = ? AND user_id = ? AND question_id = ?", scoring.ExamID, scoring.UserID, scoring.QuestionID).
		Delete(&models.ExamQuestionUserManualScoring{}).Error
	if err != nil {
		return err
	}
	return db.MasterDB.Create(scoring).Error
}

// SaveExerciseManualScoringRepository SaveExerciseManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài tập
func (r *manualScoringRepository) SaveExerciseManualScoringRepository(scoring *models.ExerciseQuestionUserManualScoring) error {
	if scoring != nil {
		t := time.Now().UTC()
		scoring.CreatedAt = &t
	}
	err := db.MasterDB.Where("exercise_id = ? AND user_id = ? AND question_id = ?", scoring.ExerciseID, scoring.UserID, scoring.QuestionID).
		Delete(&models.ExerciseQuestionUserManualScoring{}).Error
	if err != nil {
		return err
	}
	return db.MasterDB.Create(scoring).Error
}

// SaveHomeworkManualScoringRepository SaveHomeworkManualScoring lưu câu trả lời cần chấm điểm thủ công cho bài tập về nhà
func (r *manualScoringRepository) SaveHomeworkManualScoringRepository(scoring *models.HomeworkQuestionUserManualScoring) error {
	if scoring != nil {
		t := time.Now().UTC()
		scoring.CreatedAt = &t
	}
	err := db.MasterDB.Where("homework_id = ? AND user_id = ? AND question_id = ?", scoring.HomeworkID, scoring.UserID, scoring.QuestionID).
		Delete(&models.HomeworkQuestionUserManualScoring{}).Error
	if err != nil {
		return err
	}
	return db.MasterDB.Create(scoring).Error
}

func (r *saveScoreManualScoringRepository) UpdateManualScoring(examID, userID, questionID int64, score float64, scoringBy int64) error {
	now := time.Now().UTC()
	return db.MasterDB.Model(&models.ExamQuestionUserManualScoring{}).
		Where("exam_id = ? AND user_id = ? AND question_id = ?", examID, userID, questionID).
		Updates(map[string]interface{}{
			"score":      score,
			"is_scored":  true,
			"scoring_by": scoringBy,
			"scoring_at": now,
		}).Error
}

func (r *saveScoreManualScoringRepository) CheckUnscoredQuestions(examID, userID int64) (bool, error) {
	var count int64
	err := db.ReplicaDB.Model(&models.ExamQuestionUserManualScoring{}).
		Where("exam_id = ? AND user_id = ? AND is_scored = ?", examID, userID, false).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	// Trả về true nếu còn câu hỏi chưa chấm (count > 0)
	return count > 0, nil
}

func (r *saveScoreManualScoringRepository) UpdateManualScoringExercise(exerciseID, userID, questionID int64, score float64, scoringBy int64) error {
	now := time.Now().UTC()
	
	// Kiểm tra xem bản ghi có tồn tại không
	var count int64
	err := db.MasterDB.Model(&models.ExerciseQuestionUserManualScoring{}).
		Where("exercise_id = ? AND user_id = ? AND question_id = ?", exerciseID, userID, questionID).
		Count(&count).Error
	if err != nil {
		return err
	}
	
	if count > 0 {
		// Bản ghi đã tồn tại, cập nhật
		return db.MasterDB.Model(&models.ExerciseQuestionUserManualScoring{}).
			Where("exercise_id = ? AND user_id = ? AND question_id = ?", exerciseID, userID, questionID).
			Updates(map[string]interface{}{
				"score":      score,
				"is_scored":  true,
				"scoring_by": scoringBy,
				"scoring_at": now,
			}).Error
	} else {
		// Bản ghi chưa tồn tại, tạo mới
		return db.MasterDB.Create(&models.ExerciseQuestionUserManualScoring{
			ExerciseID: exerciseID,
			UserID:     userID,
			QuestionID: questionID,
			Score:      &score,
			IsScored:   true,
			ScoringBy:  &scoringBy,
			ScoringAt:  &now,
			CreatedAt:  &now,
		}).Error
	}
}

func (r *saveScoreManualScoringRepository) CheckUnscoredQuestionsExercise(exerciseID, userID int64) (bool, error) {
	var count int64
	err := db.ReplicaDB.Model(&models.ExerciseQuestionUserManualScoring{}).
		Where("exercise_id = ? AND user_id = ? AND is_scored = ?", exerciseID, userID, false).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	// Trả về true nếu còn câu hỏi chưa chấm (count > 0)
	return count > 0, nil
}

func (r *saveScoreManualScoringRepository) UpdateManualScoringHomework(homeworkID, userID, questionID int64, score float64, scoringBy int64) error {
    now := time.Now().UTC()
    var count int64
    err := db.MasterDB.Model(&models.HomeworkQuestionUserManualScoring{}).
        Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
        Count(&count).Error
    if err != nil {
        return err
    }
    if count > 0 {
        return db.MasterDB.Model(&models.HomeworkQuestionUserManualScoring{}).
            Where("homework_id = ? AND user_id = ? AND question_id = ?", homeworkID, userID, questionID).
            Updates(map[string]interface{}{
                "score":      score,
                "is_scored":  true,
                "scoring_by": scoringBy,
                "scoring_at": now,
            }).Error
    } else {
        return db.MasterDB.Create(&models.HomeworkQuestionUserManualScoring{
            HomeworkID: homeworkID,
            UserID:     userID,
            QuestionID: questionID,
            Score:      &score,
            IsScored:   true,
            ScoringBy:  &scoringBy,
            ScoringAt:  &now,
            CreatedAt:  &now,
        }).Error
    }
}

func (r *saveScoreManualScoringRepository) CheckUnscoredQuestionsHomework(homeworkID, userID int64) (bool, error) {
    var count int64
    err := db.ReplicaDB.Model(&models.HomeworkQuestionUserManualScoring{}).
        Where("homework_id = ? AND user_id = ? AND is_scored = ?", homeworkID, userID, false).
        Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (r *saveScoreManualScoringRepository) UpdateHomeworkUserManualScoringStatus(homeworkID, userID int64, hasManualScoring bool) error {
    // Dùng cast tường minh để tránh lỗi encode plan khi driver cache kiểu cũ
    return db.MasterDB.Exec(
        "UPDATE homework_users SET has_manual_scoring = ?::boolean WHERE homework_id = ? AND user_id = ?",
        hasManualScoring, homeworkID, userID,
    ).Error
}

func (r *saveScoreManualScoringRepository) UpdateExamUserManualScoringStatus(examID, userID int64, hasManualScoring bool) error {
	return db.MasterDB.Model(&models.ExamUser{}).
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Update("has_manual_scoring", hasManualScoring).Error
}

func (r *saveScoreManualScoringRepository) UpdateExerciseUserManualScoringStatus(exerciseID, userID int64, hasManualScoring bool) error {
	// Kiểm tra xem bản ghi có tồn tại không
	var count int64
	err := db.MasterDB.Model(&models.ExerciseUser{}).
		Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Count(&count).Error
	if err != nil {
		return err
	}
	
	if count > 0 {
		// Bản ghi đã tồn tại, cập nhật
		return db.MasterDB.Model(&models.ExerciseUser{}).
			Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
			Update("has_manual_scoring", hasManualScoring).Error
	} else {
		// Bản ghi chưa tồn tại, tạo mới
		now := time.Now().UTC()
		return db.MasterDB.Create(&models.ExerciseUser{
			ExerciseID:        exerciseID,
			UserID:            userID,
			HasManualScoring:  hasManualScoring,
			CreatedAt:         now,
			UpdatedAt:         now,
		}).Error
	}
}

