package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/requests"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type ExamRepository interface {
	GetAll() ([]models.Exam, error)
	GetAllWithPaging(req *requests.GetExamRequest, c *gin.Context) ([]models.Exam, int64, error)
	GetByID(id int64, c *gin.Context) (*models.Exam, error)
	IsRandomQuestion(id int64) (bool, error)
	Create(exam *models.Exam) error
	Update(exam *models.Exam) error
	Delete(id int64, deletedBy int64) error
	Assigned(ref models.ExamRefLesson) error
	AssignedLesson(id int64) (*models.Exam, error)
	UpdateEvaluate(userId, examId int64, score float64) error
}

type examRepository struct{}

func NewExamRepository() ExamRepository {
	return &examRepository{}
}

func (r *examRepository) GetAll() ([]models.Exam, error) {
	var entities []models.Exam

	err := db.ReplicaDB.
		Model(&models.Exam{}).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *examRepository) GetAllWithPaging(req *requests.GetExamRequest, c *gin.Context) ([]models.Exam, int64, error) {
	var exams []models.Exam
	var total int64

	query := db.ReplicaDB.Model(&models.Exam{}).
		Preload("ExamRefLessons").
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.exam_id = exams.id")

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

	// Count distinct exams.id because of the join
	err := query.Distinct("exams.id").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err = query.Distinct("exams.*").Order("exams.created_at DESC").Find(&exams).Error
	if err != nil {
		return nil, 0, err
	}

	return exams, total, nil
}

func (r *examRepository) GetByID(id int64, c *gin.Context) (*models.Exam, error) {
	var exam models.Exam

	query := db.ReplicaDB.Model(&models.Exam{}).Preload("ExamRefLessons")

	err := query.Where("id = ?", id).
		First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) IsRandomQuestion(id int64) (bool, error) {
	var isRandom bool
	err := db.ReplicaDB.Table("exams").
		Select("is_random_question").
		Where("id = ? AND deleted_at IS NULL", id).
		Scan(&isRandom).Error
	if err != nil {
		return false, err
	}
	return isRandom, nil
}

func (r *examRepository) Create(exam *models.Exam) error {
	return db.MasterDB.Create(exam).Error
}

func (r *examRepository) Update(exam *models.Exam) error {
	updates := map[string]interface{}{}

	if exam.Name != "" {
		updates["name"] = exam.Name
	}
	if exam.Status != 0 {
		updates["status"] = exam.Status
	}
	if exam.TimeLimit != 0 {
		updates["time_limit"] = exam.TimeLimit
	}
	if exam.MaxScore != 0 {
		updates["max_score"] = exam.MaxScore
	}
	if exam.Description != "" {
		updates["description"] = exam.Description
	}
	if exam.CoverImageInfo.Path != "" {
		updates["cover_image_info"] = exam.CoverImageInfo
	}

	updates["updated_by"] = exam.UpdatedBy
	updates["updated_at"] = time.Now()
	updates["deadline"] = exam.Deadline
	updates["object_title"] = exam.ObjectTitle
	updates["is_random_question"] = exam.IsRandomQuestion
	updates["file_infos"] = exam.FileInfos
	updates["question_form"] = exam.QuestionForm
	updates["type"] = exam.Type

	return db.MasterDB.Omit("clone_info").
		Model(&models.Exam{}).
		Where("id = ?", exam.ID).
		Updates(updates).Error
}

func (r *examRepository) Delete(id int64, deletedBy int64) error {
	model := models.Exam{}
	if err := db.MasterDB.First(&model, id).Error; err != nil {
		return err
	}

	model.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&model).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&model).Error
}


func (r *examRepository) Assigned(ref models.ExamRefLesson) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{
			{Name: "exam_id"},
			{Name: "lesson_id"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"assigned_by": ref.AssignedBy,
			"assigned_at": ref.AssignedAt,
		}),
	}).Create(&ref).Error
}

func (r *examRepository) AssignedLesson(id int64) (*models.Exam, error) {
	var exam models.Exam
	err := db.ReplicaDB.
		Preload("Lessons").
		Preload("ExamRefLessons").
		Where("id = ?", id).
		First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

func (r *examRepository) UpdateEvaluate(userId, examId int64, score float64) error {
    result := db.MasterDB.Table("exam_users").
        Where("user_id = ? AND exam_id = ?", userId, examId).
        Update("ratio", score)

    if result.Error != nil {
		config.Log.Errorf("Error updating evaluate: %v", result.Error)
        return result.Error
    }

    if result.RowsAffected == 0 {
		return errors.New("exam user not found")
    }

    return nil
}
