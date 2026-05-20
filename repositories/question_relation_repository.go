package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"

	"gorm.io/gorm"
)

type QuestionRelationRepository interface {
	BulkInsertExamQuestions(examID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertLessonPlanPartQuestions(partID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertHomeworkQuestions(homeworkID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertLevelTestQuestions(levelTestID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertExamSourceQuestions(examID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertLessonPlanPartSourceQuestions(partID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertHomeworkSourceQuestions(homeworkID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertLevelTestSourceQuestions(levelTestID int, questionScores map[int]float64, tx *gorm.DB) error

	GetQuestionIDsByHomeworkID(questionId int64) ([]int64, error)
	GetQuestionDsByExamID(questionId int64) ([]int64, error)
	GetQuestionIDsByLessonPlanPartID(questionId int64) ([]int64, error)
	GetQuestionIDsByLevelTestID(questionId int64) ([]int64, error)

	IsAssignedHomework(homeworkId, lessonId int64) (bool, error)
	IsAssignedExam(examId, lessonId int64) (bool, error)

	DeleteAllExamQuestions(examID int64, tx *gorm.DB) error
	DeleteAllHomeworkQuestions(homeworkID int64, tx *gorm.DB) error
	DeleteAllQuestionsInAssignment(examID, homeworkID int64) error

	// Exercise variants
	BulkInsertExerciseQuestions(exerciseID int, questionScores map[int]float64, tx *gorm.DB) error
	BulkInsertExerciseSourceQuestions(exerciseID int, questionScores map[int]float64, tx *gorm.DB) error
	DeleteAllExerciseQuestions(exerciseID int64, tx *gorm.DB) error
	GetQuestionIDsByExerciseID(exerciseID int64) ([]int64, error)
	IsAssignedExercise(exerciseId int64) (bool, error)
}

type questionRelationRepository struct{}

func NewQuestionRelationRepository() QuestionRelationRepository {
	return &questionRelationRepository{}
}

func (r *questionRelationRepository) BulkInsertExamQuestions(examID int, questionScores map[int]float64, tx *gorm.DB) error {
	if examID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("exam_id = ? AND is_source_question = ?", examID, false).
		Delete(&models.ExamQuestion{}).Error; err != nil {
		return err
	}

	var records []models.ExamQuestion
	for questionID, score := range questionScores {
		records = append(records, models.ExamQuestion{
			ExamID:           int64(examID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: false,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.ExamQuestion{}).
		Where("exam_id = ?", examID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Exam{}).
		Where("id = ?", examID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertLessonPlanPartQuestions(partID int, questionScores map[int]float64, tx *gorm.DB) error {
	if partID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("lesson_plan_part_id = ? AND is_source_question = ?", partID, false).
		Delete(&models.LessonPlanPartQuestion{}).Error; err != nil {
		return err
	}

	var records []models.LessonPlanPartQuestion
	for questionID, score := range questionScores {
		records = append(records, models.LessonPlanPartQuestion{
			LessonPlanPartID: int64(partID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: false,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.LessonPlanPartQuestion{}).
		Where("lesson_plan_part_id = ?", partID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.LessonPlanPart{}).
		Where("id = ?", partID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertHomeworkQuestions(homeworkID int, questionScores map[int]float64, tx *gorm.DB) error {
	if homeworkID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("homework_id = ? AND is_source_question = ?", homeworkID, false).
		Delete(&models.HomeworkQuestion{}).Error; err != nil {
		return err
	}

	var records []models.HomeworkQuestion
	for questionID, score := range questionScores {
		records = append(records, models.HomeworkQuestion{
			HomeworkID:       int64(homeworkID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: false,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.HomeworkQuestion{}).
		Where("homework_id = ?", homeworkID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Homework{}).
		Where("id = ?", homeworkID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertLevelTestQuestions(levelTestID int, questionScores map[int]float64, tx *gorm.DB) error {
	if levelTestID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("level_test_id = ? AND is_source_question = ?", levelTestID, false).
		Delete(&models.LevelTestQuestion{}).Error; err != nil {
		return err
	}

	var records []models.LevelTestQuestion
	for questionID, score := range questionScores {
		records = append(records, models.LevelTestQuestion{
			LevelTestID:      int64(levelTestID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: false,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.LevelTestQuestion{}).
		Where("level_test_id = ?", levelTestID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.LevelTest{}).
		Where("id = ?", levelTestID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertExamSourceQuestions(examID int, questionScores map[int]float64, tx *gorm.DB) error {
	if examID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("exam_id = ? AND is_source_question = ?", examID, true).
		Delete(&models.ExamQuestion{}).Error; err != nil {
		return err
	}

	var records []models.ExamQuestion
	for questionID, score := range questionScores {
		records = append(records, models.ExamQuestion{
			ExamID:           int64(examID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: true,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.ExamQuestion{}).
		Where("exam_id = ?", examID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Exam{}).
		Where("id = ?", examID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertLessonPlanPartSourceQuestions(partID int, questionScores map[int]float64, tx *gorm.DB) error {
	if partID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("lesson_plan_part_id = ? AND is_source_question = ?", partID, true).
		Delete(&models.LessonPlanPartQuestion{}).Error; err != nil {
		return err
	}

	var records []models.LessonPlanPartQuestion
	for questionID, score := range questionScores {
		records = append(records, models.LessonPlanPartQuestion{
			LessonPlanPartID: int64(partID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: true,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.LessonPlanPartQuestion{}).
		Where("lesson_plan_part_id = ?", partID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.LessonPlanPart{}).
		Where("id = ?", partID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertHomeworkSourceQuestions(homeworkID int, questionScores map[int]float64, tx *gorm.DB) error {
	if homeworkID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("homework_id = ? AND is_source_question = ?", homeworkID, true).
		Delete(&models.HomeworkQuestion{}).Error; err != nil {
		return err
	}

	var records []models.HomeworkQuestion
	for questionID, score := range questionScores {
		records = append(records, models.HomeworkQuestion{
			HomeworkID:       int64(homeworkID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: true,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.HomeworkQuestion{}).
		Where("homework_id = ?", homeworkID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Homework{}).
		Where("id = ?", homeworkID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) BulkInsertLevelTestSourceQuestions(levelTestID int, questionScores map[int]float64, tx *gorm.DB) error {
	if levelTestID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}

	if err := tx.Where("level_test_id = ? AND is_source_question = ?", levelTestID, true).
		Delete(&models.LevelTestQuestion{}).Error; err != nil {
		return err
	}

	var records []models.LevelTestQuestion
	for questionID, score := range questionScores {
		records = append(records, models.LevelTestQuestion{
			LevelTestID:      int64(levelTestID),
			QuestionID:       int64(questionID),
			Score:            score,
			IsSourceQuestion: true,
		})
	}

	if err := tx.Create(&records).Error; err != nil {
		return err
	}

	var totalScore float64
	if err := tx.Model(&models.LevelTestQuestion{}).
		Where("level_test_id = ?", levelTestID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.LevelTest{}).
		Where("id = ?", levelTestID).
		Update("max_score", totalScore).Error; err != nil {
		return err
	}

	return nil
}

func (r *questionRelationRepository) GetQuestionIDsByHomeworkID(homeworkID int64) ([]int64, error) {
	var questionIDs []int64

	err := db.MasterDB.
		Model(&models.HomeworkQuestion{}).
		Where("homework_id = ?", homeworkID).
		Pluck("question_id", &questionIDs).Error

	if err != nil {
		return nil, err
	}
	return questionIDs, nil
}

func (r *questionRelationRepository) GetQuestionDsByExamID(examID int64) ([]int64, error) {
	var questionIDs []int64

	err := db.MasterDB.
		Model(&models.ExamQuestion{}).
		Where("exam_id = ?", examID).
		Pluck("question_id", &questionIDs).Error

	if err != nil {
		return nil, err
	}
	return questionIDs, nil
}

func (r *questionRelationRepository) GetQuestionIDsByLessonPlanPartID(lessonPlanPartID int64) ([]int64, error) {
	var questionIDs []int64

	err := db.MasterDB.
		Model(&models.LessonPlanPartQuestion{}).
		Where("lesson_plan_part_id = ?", lessonPlanPartID).
		Pluck("question_id", &questionIDs).Error

	if err != nil {
		return nil, err
	}
	return questionIDs, nil
}

func (r *questionRelationRepository) GetQuestionIDsByLevelTestID(levelTestID int64) ([]int64, error) {
	var questionIDs []int64

	err := db.MasterDB.
		Model(&models.LevelTestQuestion{}).
		Where("level_test_id = ?", levelTestID).
		Pluck("question_id", &questionIDs).Error

	if err != nil {
		return nil, err
	}
	return questionIDs, nil
}

func (r *questionRelationRepository) IsAssignedHomework(homeworkId, lessonId int64) (bool, error) {
	var count int64

	query := db.MasterDB.Table("homework_ref_lessons").
		Where("homework_id = ? AND assigned_by IS NOT NULL AND assigned_by > 0", homeworkId)

	if lessonId > 0 {
		query = query.Where("lesson_id = ?", lessonId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *questionRelationRepository) IsAssignedExam(examId, lessonId int64) (bool, error) {
	var count int64

	query := db.MasterDB.Table("exam_ref_lessons").
		Where("exam_id = ? AND assigned_by IS NOT NULL AND assigned_by > 0", examId)

	if lessonId > 0 {
		query = query.Where("lesson_id = ?", lessonId)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *questionRelationRepository) DeleteAllExamQuestions(examID int64, tx *gorm.DB) error {
	if examID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}
	// Exams now use cloned_questions instead of exam_questions
	return tx.Where("assignment_id = ? AND assignment_type = ?", examID, models.ClonedQuestionTypeExam).
		Delete(&models.ClonedQuestion{}).Error
}

func (r *questionRelationRepository) DeleteAllHomeworkQuestions(homeworkID int64, tx *gorm.DB) error {
	if homeworkID == 0 {
		return nil
	}
	if tx == nil {
		tx = db.MasterDB
	}
	// Homeworks now use cloned_questions instead of homework_questions
	return tx.Where("assignment_id = ? AND assignment_type = ?", homeworkID, models.ClonedQuestionTypeHomework).
		Delete(&models.ClonedQuestion{}).Error
}

func (r *questionRelationRepository) DeleteAllQuestionsInAssignment(examID, homeworkID int64) error {
	return db.MasterDB.Transaction(func(tx *gorm.DB) error {
		if examID > 0 {
			if err := r.DeleteAllExamQuestions(examID, tx); err != nil {
				return err
			}
		}
		if homeworkID > 0 {
			if err := r.DeleteAllHomeworkQuestions(homeworkID, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *questionRelationRepository) BulkInsertExerciseQuestions(exerciseID int, questionScores map[int]float64, tx *gorm.DB) error {
	if exerciseID == 0 { return nil }
	if tx == nil { tx = db.MasterDB }
	if err := tx.Where("exercise_id = ? AND is_source_question = ?", exerciseID, false).
		Delete(&models.ExerciseQuestion{}).Error; err != nil { return err }
	var records []models.ExerciseQuestion
	for questionID, score := range questionScores {
		records = append(records, models.ExerciseQuestion{ ExerciseID: int64(exerciseID), QuestionID: int64(questionID), Score: score, IsSourceQuestion: false })
	}
	if err := tx.Create(&records).Error; err != nil { return err }
	var totalScore float64
	if err := tx.Model(&models.ExerciseQuestion{}).
		Where("exercise_id = ?", exerciseID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil { return err }
	if err := tx.Model(&models.Exercise{}).
		Where("id = ?", exerciseID).
		Update("max_score", totalScore).Error; err != nil { return err }
	return nil
}

func (r *questionRelationRepository) BulkInsertExerciseSourceQuestions(exerciseID int, questionScores map[int]float64, tx *gorm.DB) error {
	if exerciseID == 0 { return nil }
	if tx == nil { tx = db.MasterDB }
	if err := tx.Where("exercise_id = ? AND is_source_question = ?", exerciseID, true).
		Delete(&models.ExerciseQuestion{}).Error; err != nil { return err }
	var records []models.ExerciseQuestion
	for questionID, score := range questionScores {
		records = append(records, models.ExerciseQuestion{ ExerciseID: int64(exerciseID), QuestionID: int64(questionID), Score: score, IsSourceQuestion: true })
	}
	if err := tx.Create(&records).Error; err != nil { return err }
	var totalScore float64
	if err := tx.Model(&models.ExerciseQuestion{}).
		Where("exercise_id = ?", exerciseID).
		Select("COALESCE(SUM(score),0)").Scan(&totalScore).Error; err != nil { return err }
	if err := tx.Model(&models.Exercise{}).
		Where("id = ?", exerciseID).
		Update("max_score", totalScore).Error; err != nil { return err }
	return nil
}

func (r *questionRelationRepository) DeleteAllExerciseQuestions(exerciseID int64, tx *gorm.DB) error {
	if exerciseID == 0 { return nil }
	return tx.Where("exercise_id = ?", exerciseID).Delete(&models.ExerciseQuestion{}).Error
}

func (r *questionRelationRepository) GetQuestionIDsByExerciseID(exerciseID int64) ([]int64, error) {
	var questionIDs []int64
	err := db.MasterDB.Model(&models.ExerciseQuestion{}).Where("exercise_id = ?", exerciseID).Pluck("question_id", &questionIDs).Error
	if err != nil { return nil, err }
	return questionIDs, nil
}

func (r *questionRelationRepository) IsAssignedExercise(exerciseId int64) (bool, error) {
    var count int64

    err := db.MasterDB.
        Table("exercise_ref_lessons").
        Where("exercise_id = ? AND assigned_by IS NOT NULL AND assigned_by > 0", exerciseId).
        Count(&count).Error

    if err != nil {
        return false, err
    }

    return count > 0, nil
}
