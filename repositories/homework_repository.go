package repositories

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"be-cleverschool/requests"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type HomeworkRepository interface {
	GetAll() ([]models.Homework, error)
	GetAllWithPaging(req *requests.GetHomeworkRequest, c *gin.Context) ([]models.Homework, int64, error)
	GetByID(id int64, userID int64, c *gin.Context) (*dto.HomeworkDTO, error)
	IsRandomQuestion(id int64) (bool, error)
	GetStudentsDoing(req *requests.HomeworkStudentsDoingRequest) ([]dto.HomeworkStudentDoing, int64, error)
	Create(hw *models.Homework) error
	Update(hw *models.Homework) error
	Delete(id int64, deletedBy int64) error
	Assigned(ref models.HomeworkRefLesson) error
	AssignedLesson(id int64) (*models.Homework, error)
	GetTotalQuestionsFromHomeworks(homeworkID int64) (int64, error)

	UpdateEvaluate(userId, homeworkId int64, score float64) error
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
		Group(`h.id, hu.questions_completed, hu.last_question_id_completed`).
		Limit(1)

	if err := query.Scan(&homeworkDto).Error; err != nil {
		return nil, err
	}

	return &homeworkDto, nil
}

func (r *homeworkRepository) IsRandomQuestion(id int64) (bool, error) {
	var isRandom bool
	err := db.ReplicaDB.Table("homeworks").
		Select("is_random_question").
		Where("id = ? AND deleted_at IS NULL", id).
		Scan(&isRandom).Error
	if err != nil {
		return false, err
	}
	return isRandom, nil
}

func (r *homeworkRepository) GetStudentsDoing(req *requests.HomeworkStudentsDoingRequest) ([]dto.HomeworkStudentDoing, int64, error) {
	if req == nil {
		return nil, 0, errors.New("request is required")
	}

	baseQuery := db.ReplicaDB.Table("homework_users hu").
		Joins("JOIN users u ON u.id = hu.user_id AND u.deleted_at IS NULL").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("LEFT JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL").
		Where("hu.homework_id = ?", req.HomeworkID).
		Where("urr.role_id = ?", 3)

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		baseQuery = baseQuery.Where("(unaccent(u.name) ILIKE unaccent(?) OR unaccent(u.username) ILIKE unaccent(?))", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := baseQuery.Distinct("hu.user_id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := db.ReplicaDB.Table("homework_users hu").
		Select(`
			hu.homework_id,
			hu.user_id,
			u.username,
			u.name,
			COALESCE(s.name, '') AS school_name,
			COALESCE(MAX(hu.updated_at), MAX(hu.created_at)) AS submitted_at`).
		Joins("JOIN users u ON u.id = hu.user_id AND u.deleted_at IS NULL").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("LEFT JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL").
		Where("hu.homework_id = ?", req.HomeworkID).
		Where("urr.role_id = ?", 3)

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		dataQuery = dataQuery.Where("(unaccent(u.name) ILIKE unaccent(?) OR unaccent(u.username) ILIKE unaccent(?))", "%"+keyword+"%", "%"+keyword+"%")
	}

	dataQuery = dataQuery.
		Group("hu.homework_id, hu.user_id, u.name, u.username, s.name").
		Order("submitted_at DESC, u.name ASC")

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		dataQuery = dataQuery.Limit(req.Limit).Offset(offset)
	}

	var results []dto.HomeworkStudentDoing
	if err := dataQuery.Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *homeworkRepository) Create(hw *models.Homework) error {
	return db.MasterDB.Create(hw).Error
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
	model := models.Homework{}
	if err := db.MasterDB.First(&model, id).Error; err != nil {
		return err
	}

	model.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&model).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&model).Error
}

func (r *homeworkRepository) Assigned(ref models.HomeworkRefLesson) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "homework_id"},
			{Name: "lesson_id"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"assigned_by": ref.AssignedBy,
			"assigned_at": ref.AssignedAt,
		}),
	}).Create(&ref).Error
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

func (r *homeworkRepository) UpdateEvaluate(userId, homeworkId int64, score float64) error {
	result := db.MasterDB.Table("homework_users").
		Where("user_id = ? AND homework_id = ?", userId, homeworkId).
		Updates(map[string]interface{}{
			"ratio":          score,
			"status_scoring": 2,
		})

	if result.Error != nil {
		config.Log.Errorf("Error updating evaluate: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("homework user not found")
	}

	return nil
}

