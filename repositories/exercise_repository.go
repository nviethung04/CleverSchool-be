package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type ExerciseRepository interface {
    GetAll() ([]models.Exercise, error)
    GetAllWithPaging(req *requests.GetExerciseRequest, c *gin.Context) ([]models.Exercise, int64, error)
    GetByID(id int64, c *gin.Context) (*models.Exercise, error)
    Create(exercise *models.Exercise) error
    Update(exercise *models.Exercise) error
    Delete(id int64, deletedBy int64) error
    GetTotalQuestions(exerciseID int64) (int32, error)
    GetTotalQuestionsMap(exerciseIDs []int64) (map[int64]int32, error)
    Assigned(ref models.ExerciseRefLesson) error
    AssignedLesson(id int64) (*models.Exercise, error)
}

type exerciseRepository struct{}

func NewExerciseRepository() ExerciseRepository {
    return &exerciseRepository{}
}

func (r *exerciseRepository) GetAll() ([]models.Exercise, error) {
    var entities []models.Exercise

    err := db.ReplicaDB.
        Model(&models.Exercise{}).
        Order("created_at DESC").
        Find(&entities).Error

    if err != nil {
        return nil, err
    }

    return entities, nil
}

func (r *exerciseRepository) GetAllWithPaging(req *requests.GetExerciseRequest, c *gin.Context) ([]models.Exercise, int64, error) {
    var exercises []models.Exercise
    var total int64

    // Sử dụng LEFT JOIN để lấy tất cả exercises, kể cả những exercises không có liên kết với lesson
    query := db.ReplicaDB.Model(&models.Exercise{}).
		Preload("ExerciseRefLessons").
        Joins("LEFT JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id")

    if req.LessonID != nil {
        query = query.Where("erl.lesson_id = ?", *req.LessonID)
    }

    if req.IsAssigned != nil {
        if *req.IsAssigned {
            query = query.Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0")
        } else {
            query = query.Where("(erl.assigned_by IS NULL OR erl.assigned_by = 0)")
        }
    }

    // Count distinct exercises.id because of the join
    err := query.Distinct("exercises.id").Count(&total).Error
    if err != nil {
        return nil, 0, err
    }

    if req.Limit > 0 && req.Page > 0 {
        query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
    }

    err = query.Distinct("exercises.*").Order("exercises.created_at DESC").Find(&exercises).Error
    if err != nil {
        return nil, 0, err
    }

    return exercises, total, nil
}

func (r *exerciseRepository) GetByID(id int64, c *gin.Context) (*models.Exercise, error) {
    var exercise models.Exercise

    query := db.ReplicaDB.Model(&models.Exercise{}).Preload("ExerciseRefLessons")

    err := query.Where("id = ?", id).
        First(&exercise).Error
    if err != nil {
        return nil, err
    }
    return &exercise, nil
}

func (r *exerciseRepository) Create(exercise *models.Exercise) error {
	query := db.MasterDB
	if exercise.ProgramId == 0 {
		query = query.Omit("ProgramId")
	}
	return query.Create(exercise).Error
}

func (r *exerciseRepository) Update(exercise *models.Exercise) error {
    updates := map[string]interface{}{}

    if exercise.Name != "" {
        updates["name"] = exercise.Name
    }
    if exercise.Status != 0 {
        updates["status"] = exercise.Status
    }
    if exercise.TimeLimit != 0 {
        updates["time_limit"] = exercise.TimeLimit
    }
    if exercise.MaxScore != 0 {
        updates["max_score"] = exercise.MaxScore
    }
    if exercise.Description != "" {
        updates["description"] = exercise.Description
    }
    if exercise.CoverImageInfo.Path != "" {
        updates["cover_image_info"] = exercise.CoverImageInfo
    }

    updates["updated_by"] = exercise.UpdatedBy
    updates["updated_at"] = time.Now()
    updates["deadline"] = exercise.Deadline
    updates["object_title"] = exercise.ObjectTitle
	updates["is_random_question"] = exercise.IsRandomQuestion
	updates["file_infos"] = exercise.FileInfos
	updates["question_form"] = exercise.QuestionForm

    return db.MasterDB.Model(&models.Exercise{}).Where("id = ?", exercise.ID).Updates(updates).Error
}

func (r *exerciseRepository) Delete(id int64, deletedBy int64) error {
    model := models.Exercise{}
    if err := db.MasterDB.First(&model, id).Error; err != nil {
        return err
    }

    model.DeletedBy = deletedBy

    if err := db.MasterDB.Save(&model).Error; err != nil {
        return err
    }
    return db.MasterDB.Delete(&model).Error
}

func (r *exerciseRepository) GetTotalQuestions(exerciseID int64) (int32, error) {
    var total int64
    err := db.ReplicaDB.Table("cloned_questions").
        Where("assignment_id = ? AND assignment_type = ?", exerciseID, models.ClonedQuestionTypeExercise).
        Count(&total).Error
    return int32(total), err
}

func (r *exerciseRepository) GetTotalQuestionsMap(exerciseIDs []int64) (map[int64]int32, error) {
    result := make(map[int64]int32)
    if len(exerciseIDs) == 0 {
        return result, nil
    }
    type row struct{ AssignmentID int64; Cnt int32 }
    var rows []row
    err := db.ReplicaDB.Table("cloned_questions").
        Select("assignment_id, COUNT(*) as cnt").
        Where("assignment_id IN ? AND assignment_type = ?", exerciseIDs, models.ClonedQuestionTypeExercise).
        Group("assignment_id").
        Scan(&rows).Error
    if err != nil {
        return nil, err
    }
    for _, r := range rows {
        result[r.AssignmentID] = r.Cnt
    }
    return result, nil
}

func (r *exerciseRepository) Assigned(ref models.ExerciseRefLesson) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{
			{Name: "exercise_id"},
			{Name: "lesson_id"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"assigned_by": ref.AssignedBy,
			"assigned_at": ref.AssignedAt,
		}),
	}).Create(&ref).Error
}

func (r *exerciseRepository) AssignedLesson(id int64) (*models.Exercise, error) {
	var exercise models.Exercise
	err := db.ReplicaDB.
		Preload("Lessons").
		Preload("ExerciseRefLessons").
		Where("id = ?", id).
		First(&exercise).Error
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}
