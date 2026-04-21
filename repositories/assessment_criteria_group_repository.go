package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssessmentCriteriaGroupRepository interface {
	GetAllWithPaging(req *requests.GetAssessmentCriteriaGroupRequest, c *gin.Context) ([]models.AssessmentCriteriaGroup, int64, error)
	GetByID(id int64) (*models.AssessmentCriteriaGroup, error)
	Create(entity *models.AssessmentCriteriaGroup) error
	Update(entity *models.AssessmentCriteriaGroup) error
	Delete(id int64, deletedBy int64) error
	ReplaceGroupCriteria(groupID int64, criteriaIDs []int64, userID int64) error
	GetCriteriaIDsByGroupID(groupID int64) ([]int64, error)
}

type assessmentCriteriaGroupRepository struct{}

func NewAssessmentCriteriaGroupRepository() AssessmentCriteriaGroupRepository {
	return &assessmentCriteriaGroupRepository{}
}

func (r *assessmentCriteriaGroupRepository) GetAllWithPaging(req *requests.GetAssessmentCriteriaGroupRequest, c *gin.Context) ([]models.AssessmentCriteriaGroup, int64, error) {
	var items []models.AssessmentCriteriaGroup
	var total int64

	query := db.ReplicaDB.Model(&models.AssessmentCriteriaGroup{})

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("unaccent(name) ILIKE unaccent(?)", "%"+keyword+"%")
	}

	if req.Name != "" {
		query = query.Where("name ILIKE ?", "%"+strings.TrimSpace(req.Name)+"%")
	}

	if req.SubjectID != 0 {
		query = query.Where("subject_id = ?", req.SubjectID)
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

func (r *assessmentCriteriaGroupRepository) GetByID(id int64) (*models.AssessmentCriteriaGroup, error) {
	var entity models.AssessmentCriteriaGroup
	if err := db.ReplicaDB.
		Preload("RefCriteria", "deleted_at IS NULL").
		Preload("RefCriteria.AssessmentCriterion.AssessmentSubcriteria").
		Where("deleted_at IS NULL").
		First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *assessmentCriteriaGroupRepository) Create(entity *models.AssessmentCriteriaGroup) error {
	return db.MasterDB.Create(entity).Error
}

func (r *assessmentCriteriaGroupRepository) Update(entity *models.AssessmentCriteriaGroup) error {
	updates := map[string]interface{}{
		"name":       entity.Name,
		"updated_by": entity.UpdatedBy,
		"updated_at": entity.UpdatedAt,
		"has_file":   entity.HasFile,
	}

	if entity.SubjectID != 0 {
		updates["subject_id"] = entity.SubjectID
	}

	return db.MasterDB.Model(&models.AssessmentCriteriaGroup{}).Where("id = ?", entity.ID).Updates(updates).Error
}

func (r *assessmentCriteriaGroupRepository) Delete(id int64, deletedBy int64) error {
	entity := models.AssessmentCriteriaGroup{}
	if err := db.MasterDB.First(&entity, id).Error; err != nil {
		return err
	}

	entity.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&entity).Error; err != nil {
		return err
	}

	return db.MasterDB.Delete(&entity).Error
}

func (r *assessmentCriteriaGroupRepository) ReplaceGroupCriteria(groupID int64, criteriaIDs []int64, userID int64) error {
	return db.MasterDB.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		// Soft delete các criteria cũ với UTC
		if err := tx.Model(&models.AssessmentCriteriaGroupRefCriterion{}).
			Where("assessment_criteria_group_id = ? AND deleted_at IS NULL", groupID).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"deleted_by": userID,
				"updated_at": now,
				"updated_by": userID,
			}).Error; err != nil {
			return err
		}

		if len(criteriaIDs) == 0 {
			return nil
		}

		refs := make([]models.AssessmentCriteriaGroupRefCriterion, 0, len(criteriaIDs))
		for _, criterionID := range criteriaIDs {
			if criterionID == 0 {
				continue
			}
			refs = append(refs, models.AssessmentCriteriaGroupRefCriterion{
				AssessmentCriteriaGroupID: groupID,
				AssessmentCriteriaID:      criterionID,
				CreatedAt:                 now,
				CreatedBy:                 userID,
				UpdatedAt:                 now,
				UpdatedBy:                 userID,
			})
		}

		if len(refs) == 0 {
			return nil
		}

		return tx.Create(&refs).Error
	})
}

func (r *assessmentCriteriaGroupRepository) GetCriteriaIDsByGroupID(groupID int64) ([]int64, error) {
	var refs []models.AssessmentCriteriaGroupRefCriterion
	if err := db.ReplicaDB.
		Where("assessment_criteria_group_id = ? AND deleted_at IS NULL", groupID).
		Find(&refs).Error; err != nil {
		return nil, err
	}

	criteriaIDs := make([]int64, 0, len(refs))
	for _, ref := range refs {
		if ref.AssessmentCriteriaID > 0 {
			criteriaIDs = append(criteriaIDs, ref.AssessmentCriteriaID)
		}
	}

	return criteriaIDs, nil
}
