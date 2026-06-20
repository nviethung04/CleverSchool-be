package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/requests"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomeworkRepository interface {
	GetAll() ([]models.Homework, error)
	GetAllWithPaging(req *requests.GetHomeworkRequest, c *gin.Context) ([]models.Homework, int64, error)
	GetByID(id int64, userID int64, c *gin.Context) (*dto.HomeworkDTO, error)
	Create(hw *models.Homework) error
	Update(hw *models.Homework) error
	Delete(id int64, deletedBy int64) error
	Assigned(ref models.HomeworkRefLesson) error
	AssignedLesson(id int64) (*models.Homework, error)
    GetTotalQuestionsFromHomeworks(homeworkID int64) (int64, error)
}

type homeworkRepository struct{}

func NewHomeworkRepository() HomeworkRepository {
	return &homeworkRepository{}
}

func (r *homeworkRepository) GetAll() ([]models.Homework, error) {
	var entities []models.Homework

	err := db.ReplicaDB.
		Model(&models.Homework{}).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *homeworkRepository) GetAllWithPaging(req *requests.GetHomeworkRequest, c *gin.Context) ([]models.Homework, int64, error) {
	var homeworks []models.Homework
	var total int64

	query := db.ReplicaDB.Model(&models.Homework{}).
		Preload("HomeworkRefLessons").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id")

	if req.LessonID != nil {
		query = query.Where("hrl.lesson_id = ?", *req.LessonID)
	}

	if req.IsAssigned != nil {
		if *req.IsAssigned {
			query = query.Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0")
		} else {
			query = query.Where("hrl.assigned_by IS NULL OR hrl.assigned_by = 0")
		}
	}

	keyword := strings.TrimSpace(req.Keyword)

	if keyword != "" {
		query = query.Where("unaccent(homeworks.name) ILIKE unaccent(?) OR unaccent(homeworks.description) ILIKE unaccent(?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Đếm distinct để tránh đếm trùng do join
	if err := query.Distinct("homeworks.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err := query.Distinct("homeworks.*").Order("homeworks.created_at DESC").Find(&homeworks).Error
	if err != nil {
		return nil, 0, err
	}

	return homeworks, total, nil
}

func (r *homeworkRepository) GetByID(id int64, userID int64, c *gin.Context) (*dto.HomeworkDTO, error) {
	var homeworkDto dto.HomeworkDTO

	query := db.ReplicaDB.Table("homeworks h").
		Select(`
			h.*,
			h.total_questions as total_question,
			COALESCE(hu.questions_completed, 0) as question_completed,
			COALESCE(hu.last_question_id_completed, 0) as last_question_id_completed,
			CASE
				WHEN MAX(CASE WHEN hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 THEN 1 ELSE 0 END) = 1
				THEN true ELSE false END as is_assigned
		`).
		Joins(`LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ?`, userID).
		Joins(`LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id`).
		Where("h.id = ? AND h.deleted_at IS NULL", id).
		Group(`h.id, hu.questions_completed, hu.last_question_id_completed`)

	if err := query.Scan(&homeworkDto).Error; err != nil {
		return nil, err
	}

	return &homeworkDto, nil
}

func (r *homeworkRepository) Create(hw *models.Homework) error {
	query := db.MasterDB
	if hw.ProgramId == 0 {
		query = query.Omit("ProgramId")
	}
	return query.Create(hw).Error
}

func (r *homeworkRepository) Update(hw *models.Homework) error {
	updates := map[string]interface{}{}

	if hw.Name != "" {
		updates["name"] = hw.Name
	}
	if hw.Status != 0 {
		updates["status"] = hw.Status
	}
	if hw.Description != "" {
		updates["description"] = hw.Description
	}
	if hw.MaxScore != 0 {
		updates["max_score"] = hw.MaxScore
	}
	if hw.CoverImageInfo.Path != "" {
		updates["cover_image_info"] = hw.CoverImageInfo
	}

	updates["updated_by"] = hw.UpdatedBy
	updates["updated_at"] = time.Now().UTC()
	updates["object_title"] = hw.ObjectTitle
	updates["is_random_question"] = hw.IsRandomQuestion
	updates["file_infos"] = hw.FileInfos
	updates["question_form"] = hw.QuestionForm

	return db.MasterDB.Omit("clone_info").
		Model(&models.Homework{}).
		Where("id = ?", hw.ID).
		Updates(updates).Error
}

func (r *homeworkRepository) Delete(id int64, deletedBy int64) error {
	if err := db.MasterDB.Model(&models.Homework{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&models.Homework{}, id).Error
}



func (r *homeworkRepository) Assigned(ref models.HomeworkRefLesson) error {
	isAssign := ref.AssignedBy != nil && *ref.AssignedBy > 0

	var existing models.HomeworkRefLesson
	query := db.MasterDB.Where(
		"lesson_id = ? AND homework_id = ?",
		ref.LessonId, ref.HomeworkId,
	)
	if ref.CourseId == 0 {
		query = query.Where("course_id IS NULL OR course_id = 0")
	} else {
		query = query.Where("course_id = ?", ref.CourseId)
	}
	err := query.First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !isAssign {
			return nil
		}
		return db.MasterDB.Create(&ref).Error
	}
	if err != nil {
		return err
	}

	if !isAssign {
		return db.MasterDB.Exec(
			"UPDATE homework_ref_lessons SET assigned_by = NULL, assigned_at = NULL WHERE id = ?",
			existing.ID,
		).Error
	}

	return db.MasterDB.Model(&existing).Updates(map[string]interface{}{
		"assigned_by": ref.AssignedBy,
		"assigned_at": ref.AssignedAt,
	}).Error
}

func (r *homeworkRepository) AssignedLesson(id int64) (*models.Homework, error) {
	var homework models.Homework
	err := db.ReplicaDB.
		Preload("Lessons").
		Preload("HomeworkRefLessons").
		Where("id = ?", id).
		First(&homework).Error
	if err != nil {
		return nil, err
	}
	return &homework, nil
}

func (r *homeworkRepository) GetTotalQuestionsFromHomeworks(homeworkID int64) (int64, error) {
    var total int64
    err := db.ReplicaDB.Table("homeworks").
        Select("COALESCE(total_questions,0)").
        Where("id = ?", homeworkID).
        Scan(&total).Error
    return total, err
}
