package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentCriterionRepository interface {
	GetAllWithPaging(req *requests.GetAssessmentCriterionRequest, c *gin.Context) ([]models.AssessmentCriterion, int64, error)
	GetByID(id int64) (*models.AssessmentCriterion, error)
	Create(entity *models.AssessmentCriterion) error
	Update(entity *models.AssessmentCriterion) error
	Delete(id int64, deletedBy int64) error
	CreateRefCriterion(ref *models.AssessmentRefCriterion) error
	GetRefCriteriaIDsByAssessmentID(assessmentID int64) ([]int64, error)
	SoftDeleteRefCriteriaByAssessmentID(assessmentID int64, deletedBy int64) error
	RestoreRefCriterion(assessmentID, criterionID int64, updatedBy int64) error
}

type assessmentCriterionRepository struct{}

func NewAssessmentCriterionRepository() AssessmentCriterionRepository {
	return &assessmentCriterionRepository{}
}

func (r *assessmentCriterionRepository) GetAllWithPaging(req *requests.GetAssessmentCriterionRequest, c *gin.Context) ([]models.AssessmentCriterion, int64, error) {
	var items []models.AssessmentCriterion
	var total int64

	query := db.ReplicaDB.Model(&models.AssessmentCriterion{}).Preload("AssessmentSubcriteria")

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("unaccent(name) ILIKE unaccent(?) OR unaccent(description) ILIKE unaccent(?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	if req.SubjectID != nil {
		query = query.Where("subject_id = ?", *req.SubjectID)
	}

	if req.HasSubcriteria != nil {
		query = query.Where("has_subcriteria = ?", *req.HasSubcriteria)
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

func (r *assessmentCriterionRepository) GetByID(id int64) (*models.AssessmentCriterion, error) {
	var entity models.AssessmentCriterion
	if err := db.ReplicaDB.Preload("AssessmentSubcriteria").First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *assessmentCriterionRepository) Create(entity *models.AssessmentCriterion) error {
	return db.MasterDB.Create(entity).Error
}

func (r *assessmentCriterionRepository) Update(entity *models.AssessmentCriterion) error {
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
	if entity.SubjectID != 0 {
		updates["subject_id"] = entity.SubjectID
	}
	updates["has_subcriteria"] = entity.HasSubcriteria
	updates["max_score"] = entity.MaxScore

	return db.MasterDB.Model(&models.AssessmentCriterion{}).Where("id = ?", entity.ID).Updates(updates).Error
}

func (r *assessmentCriterionRepository) Delete(id int64, deletedBy int64) error {
	entity := models.AssessmentCriterion{}
	if err := db.MasterDB.First(&entity, id).Error; err != nil {
		return err
	}

	entity.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&entity).Error; err != nil {
		return err
	}

	return db.MasterDB.Delete(&entity).Error
}

func (r *assessmentCriterionRepository) CreateRefCriterion(ref *models.AssessmentRefCriterion) error {
	return db.MasterDB.Create(ref).Error
}

func (r *assessmentCriterionRepository) GetRefCriteriaIDsByAssessmentID(assessmentID int64) ([]int64, error) {
	var refs []models.AssessmentRefCriterion
	if err := db.ReplicaDB.
		Where("assessment_id = ? AND deleted_at IS NULL", assessmentID).
		Find(&refs).Error; err != nil {
		return nil, err
	}

	criteriaIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		if ref.AssessmentCriterionID > 0 {
			criteriaIDs = append(criteriaIDs, ref.AssessmentCriterionID)
		}
	}

	return criteriaIDs, nil
}

func (r *assessmentCriterionRepository) SoftDeleteRefCriteriaByAssessmentID(assessmentID int64, deletedBy int64) error {
	now := time.Now().UTC()
	return db.MasterDB.Model(&models.AssessmentRefCriterion{}).
		Where("assessment_id = ? AND deleted_at IS NULL", assessmentID).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
		}).Error
}

func (r *assessmentCriterionRepository) RestoreRefCriterion(assessmentID, criterionID int64, updatedBy int64) error {
	now := time.Now().UTC()
	return db.MasterDB.Model(&models.AssessmentRefCriterion{}).
		Where("assessment_id = ? AND assessment_criterion_id = ?", assessmentID, criterionID).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"deleted_by": 0,
			"updated_by": updatedBy,
			"updated_at": now,
		}).Error
}
