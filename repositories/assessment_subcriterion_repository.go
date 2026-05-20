package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/requests"
	"strings"

	"github.com/gin-gonic/gin"
)

type AssessmentSubcriterionRepository interface {
	GetAllWithPaging(req *requests.GetAssessmentSubcriterionRequest, c *gin.Context) ([]models.AssessmentSubcriterion, int64, error)
	GetByID(id int64) (*models.AssessmentSubcriterion, error)
	Create(entity *models.AssessmentSubcriterion) error
	Update(entity *models.AssessmentSubcriterion) error
	Delete(id int64, deletedBy int64) error
	GetIDsByCriterionID(criterionID int64) ([]int64, error)
}

type assessmentSubcriterionRepository struct{}

func NewAssessmentSubcriterionRepository() AssessmentSubcriterionRepository {
	return &assessmentSubcriterionRepository{}
}

func (r *assessmentSubcriterionRepository) GetAllWithPaging(req *requests.GetAssessmentSubcriterionRequest, c *gin.Context) ([]models.AssessmentSubcriterion, int64, error) {
	var items []models.AssessmentSubcriterion
	var total int64

	query := db.ReplicaDB.Model(&models.AssessmentSubcriterion{})

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("unaccent(name) ILIKE unaccent(?) OR unaccent(description) ILIKE unaccent(?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	if req.CriterionID != nil {
		query = query.Where("criterion_id = ?", *req.CriterionID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *assessmentSubcriterionRepository) GetByID(id int64) (*models.AssessmentSubcriterion, error) {
	var entity models.AssessmentSubcriterion
	if err := db.ReplicaDB.First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *assessmentSubcriterionRepository) Create(entity *models.AssessmentSubcriterion) error {
	return db.MasterDB.Create(entity).Error
}

func (r *assessmentSubcriterionRepository) Update(entity *models.AssessmentSubcriterion) error {
	updates := map[string]interface{}{
		"updated_by": entity.UpdatedBy,
		"updated_at": entity.UpdatedAt,
	}

	if entity.Name != "" {
		updates["name"] = entity.Name
	}
	if entity.Description != "" {
		updates["description"] = entity.Description
	}
	if entity.CriterionID != 0 {
		updates["criterion_id"] = entity.CriterionID
	}
	updates["max_score"] = entity.MaxScore

	return db.MasterDB.Model(&models.AssessmentSubcriterion{}).Where("id = ?", entity.ID).Updates(updates).Error
}

func (r *assessmentSubcriterionRepository) Delete(id int64, deletedBy int64) error {
	entity := models.AssessmentSubcriterion{}
	if err := db.MasterDB.First(&entity, id).Error; err != nil {
		return err
	}

	entity.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&entity).Error; err != nil {
		return err
	}

	return db.MasterDB.Delete(&entity).Error
}

func (r *assessmentSubcriterionRepository) GetIDsByCriterionID(criterionID int64) ([]int64, error) {
	var subcriteria []models.AssessmentSubcriterion
	err := db.ReplicaDB.
		Where("criterion_id = ? AND deleted_at IS NULL", criterionID).
		Find(&subcriteria).Error
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(subcriteria))
	for _, sub := range subcriteria {
		ids = append(ids, sub.ID)
	}
	return ids, nil
}

