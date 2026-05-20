package repositories

import (
	"fmt"
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/requests"
	"strings"
	"time"
)

type FeedbackRepository interface {
	GetAll() ([]dto.FeedbackResponse, error)
	GetAllWithPaging(req *requests.GetFeedbackRequest) ([]dto.FeedbackResponse, int64, error)
	GetByID(id int64) (*dto.FeedbackResponse, error)
	Create(feedback *models.Feedback) error
	Update(feedback *models.Feedback) error
	UpdatePartial(id int64, updates map[string]interface{}) error
	Delete(id int64, deletedBy int64) error
}

type feedbackRepository struct{}

func NewFeedbackRepository() FeedbackRepository {
	return &feedbackRepository{}
}

func (r *feedbackRepository) GetAll() ([]dto.FeedbackResponse, error) {
	var entities []dto.FeedbackResponse

	err := db.ReplicaDB.
		Table("feedbacks").
		Select("feedbacks.*, users.username, users.name, roles.name as role_name").
		Joins("LEFT JOIN users ON feedbacks.user_id = users.id").
		Joins("LEFT JOIN roles ON feedbacks.role_id = roles.id").
		Order("feedbacks.created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (r *feedbackRepository) GetAllWithPaging(req *requests.GetFeedbackRequest) ([]dto.FeedbackResponse, int64, error) {
	var feedbacks []dto.FeedbackResponse
	var total int64

	query := db.ReplicaDB.
		Table("feedbacks").
		Select("feedbacks.*, users.username, users.name, roles.name as role_name").
		Joins("LEFT JOIN users ON feedbacks.user_id = users.id").
		Joins("LEFT JOIN roles ON feedbacks.role_id = roles.id")

	if req.UserID != nil {
		query = query.Where("feedbacks.user_id = ?", req.UserID)
	}

	if req.RoleID != nil {
		query = query.Where("feedbacks.role_id = ?", req.RoleID)
	}

	if req.Status != nil {
		query = query.Where("feedbacks.status = ?", req.Status)
	}

	if req.Type != nil {
		query = query.Where("feedbacks.type = ?", req.Type)
	}

	// Filter theo ngày (chỉ so sánh ngày, không so sánh giờ)
	if req.StartDate != nil {
		startTime := time.Unix(*req.StartDate, 0)
		startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
		query = query.Where("DATE(feedbacks.created_at) >= ?", startDate.Format("2006-01-02"))
	}

	if req.EndDate != nil {
		endTime := time.Unix(*req.EndDate, 0)
		endDate := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
		query = query.Where("DATE(feedbacks.created_at) <= ?", endDate.Format("2006-01-02"))
	}

	// Validate start_date <= end_date
	if req.StartDate != nil && req.EndDate != nil && *req.StartDate > *req.EndDate {
		return nil, 0, fmt.Errorf("start_date cannot be greater than end_date")
	}

	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		query = query.Where("unaccent(feedbacks.content) ILIKE unaccent(?)", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	// Order by
	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	switch orderBy {
	case "oldest":
		query = query.Order("feedbacks.created_at ASC")
	case "newest":
		fallthrough
	default:
		query = query.Order("feedbacks.created_at DESC")
	}

	err := query.Find(&feedbacks).Error
	if err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

func (r *feedbackRepository) GetByID(id int64) (*dto.FeedbackResponse, error) {
	var feedback dto.FeedbackResponse
	err := db.ReplicaDB.
		Table("feedbacks").
		Select("feedbacks.*, users.username, users.name, roles.name as role_name").
		Joins("LEFT JOIN users ON feedbacks.user_id = users.id").
		Joins("LEFT JOIN roles ON feedbacks.role_id = roles.id").
		Where("feedbacks.id = ?", id).
		First(&feedback).Error
	
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *feedbackRepository) Create(feedback *models.Feedback) error {
	return db.MasterDB.Create(feedback).Error
}

func (r *feedbackRepository) Update(feedback *models.Feedback) error {
	updates := map[string]interface{}{}

	// Chỉ cập nhật những trường được truyền vào
	if feedback.Content != "" {
		updates["content"] = feedback.Content
	}
	
	if feedback.Title != "" {
		updates["title"] = feedback.Title
	}
	
	if feedback.FileInfo.Path != "" {
		updates["file_info"] = feedback.FileInfo
	}
	
	if feedback.Status != nil {
		updates["status"] = *feedback.Status
	}
	
	if feedback.Response != "" {
		updates["response"] = feedback.Response
	}
	
	if feedback.Note != "" {
		updates["note"] = feedback.Note
	}
	
	if feedback.Type != nil {
		updates["type"] = *feedback.Type
	}

	updates["updated_by"] = feedback.UpdatedBy
	updates["updated_at"] = time.Now().UTC()

	return db.MasterDB.Model(&models.Feedback{}).Where("id = ?", feedback.ID).Updates(updates).Error
}

func (r *feedbackRepository) UpdatePartial(id int64, updates map[string]interface{}) error {
	return db.MasterDB.Model(&models.Feedback{}).Where("id = ?", id).Updates(updates).Error
}

func (r *feedbackRepository) Delete(id int64, deletedBy int64) error {
	model := models.Feedback{}
	if err := db.MasterDB.First(&model, id).Error; err != nil {
		return err
	}

	model.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&model).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&model).Error
} 