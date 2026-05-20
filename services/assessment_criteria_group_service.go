package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentCriteriaGroupService interface {
	GetAll(c *gin.Context) ([]models.AssessmentCriteriaGroup, int64, error)
	GetByID(c *gin.Context, id int) (*prot.AssessmentCriteriaGroup, error)
	Create(c *gin.Context, req *prot.AssessmentCriteriaGroupRequest) (*models.AssessmentCriteriaGroup, error)
	Update(c *gin.Context, req *prot.AssessmentCriteriaGroupRequest) (*models.AssessmentCriteriaGroup, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.AssessmentCriteriaGroup, error)
	CreateWithCriteria(c *gin.Context, req *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest) (*prot.AssessmentCriteriaGroup, error)
	UpdateWithCriteria(c *gin.Context, id int, req *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest) (*prot.AssessmentCriteriaGroup, error)
}

type assessmentCriteriaGroupService struct {
	repo             repositories.AssessmentCriteriaGroupRepository
	criterionRepo    repositories.AssessmentCriterionRepository
	subcriterionRepo repositories.AssessmentSubcriterionRepository
	resource         resources.AssessmentCriteriaGroupResource
}

func NewAssessmentCriteriaGroupService(repo repositories.AssessmentCriteriaGroupRepository) AssessmentCriteriaGroupService {
	return &assessmentCriteriaGroupService{
		repo:             repo,
		criterionRepo:    repositories.NewAssessmentCriterionRepository(),
		subcriterionRepo: repositories.NewAssessmentSubcriterionRepository(),
		resource:         resources.NewAssessmentCriteriaGroupResource(),
	}
}

func (s *assessmentCriteriaGroupService) GetAll(c *gin.Context) ([]models.AssessmentCriteriaGroup, int64, error) {
	var req requests.GetAssessmentCriteriaGroupRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	return s.repo.GetAllWithPaging(&req, c)
}

func (s *assessmentCriteriaGroupService) GetByID(c *gin.Context, id int) (*prot.AssessmentCriteriaGroup, error) {
	group, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}
	return s.resource.FormatGroup(group), nil
}

func (s *assessmentCriteriaGroupService) Create(c *gin.Context, req *prot.AssessmentCriteriaGroupRequest) (*models.AssessmentCriteriaGroup, error) {
	// Đã truyền copy_from_assessment_criteria_group_id → chỉ dùng logic copy
	copyFromID := req.GetCopyFromAssessmentCriteriaGroupId()
	if copyFromID == 0 {
		copyFromID = getCopyFromAssessmentCriteriaGroupIDFromBody(c)
	}
	if copyFromID > 0 {
		return s.createByCopyFromGroup(c, copyFromID)
	}

	entity := s.resource.FormatModelGroup(req)
	if entity == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	entity.CreatedBy = int64(utils.GetCurrentUserId(c))
	entity.CreatedAt = now
	entity.UpdatedAt = now

	if err := s.repo.Create(entity); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceGroupCriteria(entity.ID, req.CriteriaIds, entity.CreatedBy); err != nil {
		return nil, err
	}

	return entity, nil
}

func getCopyFromAssessmentCriteriaGroupIDFromBody(c *gin.Context) int64 {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return 0
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return 0
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return 0
	}
	if v, ok := m["copy_from_assessment_criteria_group_id"]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int:
			return int64(n)
		case int64:
			return n
		}
	}
	return 0
}

// createByCopyFromGroup: tạo group mới copy từ group nguồn; copy từng criterion trong assessment_criteria rồi gán criteria mới vào group mới qua assessment_criteria_group_ref_criteria.
func (s *assessmentCriteriaGroupService) createByCopyFromGroup(c *gin.Context, sourceID int64) (*models.AssessmentCriteriaGroup, error) {
	var sourceRow models.AssessmentCriteriaGroup
	if err := db.ReplicaDB.Table("assessment_criteria_groups").Where("id = ? AND deleted_at IS NULL", sourceID).First(&sourceRow).Error; err != nil {
		return nil, fmt.Errorf("không tìm thấy assessment criteria group id %d để copy: %w", sourceID, err)
	}

	now := time.Now().UTC()
	userID := int64(utils.GetCurrentUserId(c))

	newRow := &models.AssessmentCriteriaGroup{
		Name:      sourceRow.Name,
		SubjectID: sourceRow.SubjectID,
		HasFile:   sourceRow.HasFile,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	if err := s.repo.Create(newRow); err != nil {
		return nil, err
	}

	criteriaIDs, err := s.repo.GetCriteriaIDsByGroupID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("lấy criteria từ group nguồn: %w", err)
	}

	var newCriteriaIDs []int64
	for _, oldCriterionID := range criteriaIDs {
		if oldCriterionID == 0 {
			continue
		}
		oldCriterion, err := s.criterionRepo.GetByID(oldCriterionID)
		if err != nil || oldCriterion == nil {
			continue
		}
		// Sao chép hàng trong assessment_criteria (trường dữ liệu giống hàng cũ, audit mới)
		newCriterion := &models.AssessmentCriterion{
			SubjectID:      oldCriterion.SubjectID,
			Name:           oldCriterion.Name,
			Description:    oldCriterion.Description,
			HasSubcriteria: oldCriterion.HasSubcriteria,
			MaxScore:       oldCriterion.MaxScore,
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedBy:      userID,
			UpdatedBy:      userID,
		}
		if err := s.criterionRepo.Create(newCriterion); err != nil {
			return nil, fmt.Errorf("copy assessment_criterion id %d: %w", oldCriterionID, err)
		}
		// Copy subcriteria nếu có (criterion_id trỏ sang id criterion mới)
		for _, sub := range oldCriterion.AssessmentSubcriteria {
			newSub := &models.AssessmentSubcriterion{
				CriterionID: newCriterion.ID,
				Name:        sub.Name,
				Description: sub.Description,
				MaxScore:    sub.MaxScore,
				CreatedAt:   now,
				UpdatedAt:   now,
				CreatedBy:   userID,
				UpdatedBy:   userID,
			}
			if err := s.subcriterionRepo.Create(newSub); err != nil {
				return nil, fmt.Errorf("copy assessment_subcriterion: %w", err)
			}
		}
		newCriteriaIDs = append(newCriteriaIDs, newCriterion.ID)
	}

	if len(newCriteriaIDs) > 0 {
		if err := s.repo.ReplaceGroupCriteria(newRow.ID, newCriteriaIDs, userID); err != nil {
			return nil, fmt.Errorf("gán criteria mới vào group: %w", err)
		}
	}

	return newRow, nil
}

func (s *assessmentCriteriaGroupService) Update(c *gin.Context, req *prot.AssessmentCriteriaGroupRequest) (*models.AssessmentCriteriaGroup, error) {
	entity := s.resource.FormatModelGroup(req)
	if entity == nil {
		return nil, nil
	}

	entity.UpdatedBy = int64(utils.GetCurrentUserId(c))
	entity.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(entity); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceGroupCriteria(entity.ID, req.CriteriaIds, entity.UpdatedBy); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *assessmentCriteriaGroupService) Delete(c *gin.Context, id int) error {
	return s.repo.Delete(int64(id), int64(utils.GetCurrentUserId(c)))
}

func (s *assessmentCriteriaGroupService) Restore(c *gin.Context, id int) (*models.AssessmentCriteriaGroup, error) {
	baseRepo := base.NewBaseRepository[*models.AssessmentCriteriaGroup]()
	baseRepo.SetContext(c)
	item, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *item, nil
}

func (s *assessmentCriteriaGroupService) CreateWithCriteria(c *gin.Context, req *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest) (*prot.AssessmentCriteriaGroup, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	// Nếu truyền copy_from_assessment_criteria_group_id thì chỉ dùng logic copy (bỏ qua group/criteria trong body)
	copyFromID := int64(0)
	if req.Group != nil {
		copyFromID = req.Group.GetCopyFromAssessmentCriteriaGroupId()
	}
	if copyFromID == 0 {
		copyFromID = getCopyFromAssessmentCriteriaGroupIDFromBody(c)
	}
	if copyFromID > 0 {
		newRow, err := s.createByCopyFromGroup(c, copyFromID)
		if err != nil {
			return nil, err
		}
		createdGroup, err := s.repo.GetByID(newRow.ID)
		if err != nil {
			return nil, err
		}
		return s.resource.FormatGroup(createdGroup), nil
	}

	if req.Group == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.Group.SubjectId == 0 {
		return nil, fmt.Errorf("subject_id is required")
	}

	userID := int64(utils.GetCurrentUserId(c))
	now := time.Now().UTC()

	// Tạo group
	group := &models.AssessmentCriteriaGroup{
		Name:      req.Group.Name,
		SubjectID: req.Group.SubjectId,
		HasFile:   req.Group.HasFile,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedBy: userID,
		UpdatedAt: now,
	}

	if err := s.repo.Create(group); err != nil {
		return nil, err
	}

	// Tạo criteria và subcriteria
	var criteriaIDs []int64
	for _, criterionInput := range req.Criteria {
		hasSubcriteria := len(criterionInput.Subcriteria) > 0
		maxScore := criterionInput.MaxScore
		if hasSubcriteria {
			maxScore = s.sumSubcriteriaInputs(criterionInput.Subcriteria)
		}

		criterion := &models.AssessmentCriterion{
			SubjectID:      req.Group.SubjectId,
			Name:           criterionInput.Name,
			Description:    criterionInput.Description,
			HasSubcriteria: hasSubcriteria,
			MaxScore:       maxScore,
			CreatedBy:      userID,
			CreatedAt:      now,
			UpdatedBy:      userID,
			UpdatedAt:      now,
		}

		if err := s.criterionRepo.Create(criterion); err != nil {
			return nil, err
		}

		// Tạo subcriteria nếu có
		if hasSubcriteria {
			for _, subInput := range criterionInput.Subcriteria {
				subcriterion := &models.AssessmentSubcriterion{
					CriterionID: criterion.ID,
					Name:        subInput.Name,
					Description: subInput.Description,
					MaxScore:    subInput.MaxScore,
					CreatedBy:   userID,
					CreatedAt:   now,
					UpdatedBy:   userID,
					UpdatedAt:   now,
				}

				if err := s.subcriterionRepo.Create(subcriterion); err != nil {
					return nil, err
				}
			}
		}

		criteriaIDs = append(criteriaIDs, criterion.ID)
	}

	// Gán criteria vào group
	if err := s.repo.ReplaceGroupCriteria(group.ID, criteriaIDs, userID); err != nil {
		return nil, err
	}

	// Lấy lại group để trả về
	createdGroup, err := s.repo.GetByID(group.ID)
	if err != nil {
		return nil, err
	}

	return s.resource.FormatGroup(createdGroup), nil
}

func (s *assessmentCriteriaGroupService) UpdateWithCriteria(c *gin.Context, id int, req *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest) (*prot.AssessmentCriteriaGroup, error) {
	if req == nil || req.Group == nil {
		return nil, fmt.Errorf("request is required")
	}

	// Kiểm tra group tồn tại
	existingGroup, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	userID := int64(utils.GetCurrentUserId(c))
	now := time.Now().UTC()

	// Update group info
	existingGroup.Name = req.Group.Name
	if req.Group.SubjectId > 0 {
		existingGroup.SubjectID = req.Group.SubjectId
	}
	existingGroup.HasFile = req.Group.HasFile
	existingGroup.UpdatedBy = userID
	existingGroup.UpdatedAt = now

	if err := s.repo.Update(existingGroup); err != nil {
		return nil, err
	}

	// Lấy danh sách criteria hiện tại của group
	currentCriteriaIDs, err := s.repo.GetCriteriaIDsByGroupID(int64(id))
	if err != nil {
		return nil, err
	}

	// Xử lý criteria: update/create/delete
	var newCriteriaIDs []int64
	for _, criterionInput := range req.Criteria {
		var criterionID int64

		if criterionInput.Id > 0 {
			// Update existing criterion
			criterion, err := s.criterionRepo.GetByID(criterionInput.Id)
			if err != nil {
				return nil, fmt.Errorf("criterion %d not found: %w", criterionInput.Id, err)
			}

			hasSubcriteria := len(criterionInput.Subcriteria) > 0
			maxScore := criterionInput.MaxScore
			if hasSubcriteria {
				maxScore = s.sumSubcriteriaInputs(criterionInput.Subcriteria)
			}

			criterion.Name = criterionInput.Name
			criterion.Description = criterionInput.Description
			criterion.HasSubcriteria = hasSubcriteria
			criterion.MaxScore = maxScore
			criterion.UpdatedBy = userID
			criterion.UpdatedAt = now

			if err := s.criterionRepo.Update(criterion); err != nil {
				return nil, err
			}

			criterionID = criterion.ID

			// Xử lý subcriteria
			if hasSubcriteria {
				// Lấy danh sách subcriteria hiện tại
				var currentSubcriteria []models.AssessmentSubcriterion
				if err := db.MasterDB.Where("criterion_id = ? AND deleted_at IS NULL", criterionID).
					Find(&currentSubcriteria).Error; err != nil {
					return nil, err
				}

				currentSubMap := make(map[int64]bool)
				for _, sub := range currentSubcriteria {
					currentSubMap[sub.ID] = true
				}

				// Update/create subcriteria
				for _, subInput := range criterionInput.Subcriteria {
					if subInput.Id > 0 {
						// Update existing
						subcriterion := &models.AssessmentSubcriterion{
							ID:          subInput.Id,
							CriterionID:  criterionID,
							Name:         subInput.Name,
							Description:  subInput.Description,
							MaxScore:     subInput.MaxScore,
							UpdatedBy:    userID,
							UpdatedAt:    now,
						}
						if err := s.subcriterionRepo.Update(subcriterion); err != nil {
							return nil, err
						}
						delete(currentSubMap, subInput.Id)
					} else {
						// Create new
						subcriterion := &models.AssessmentSubcriterion{
							CriterionID:  criterionID,
							Name:         subInput.Name,
							Description:  subInput.Description,
							MaxScore:     subInput.MaxScore,
							CreatedBy:    userID,
							CreatedAt:    now,
							UpdatedBy:    userID,
							UpdatedAt:    now,
						}
						if err := s.subcriterionRepo.Create(subcriterion); err != nil {
							return nil, err
						}
					}
				}

				// Soft delete subcriteria không còn trong request
				for subID := range currentSubMap {
					if err := s.subcriterionRepo.Delete(subID, userID); err != nil {
						return nil, err
					}
				}
			} else {
				// Nếu không có subcriteria nữa, soft delete tất cả subcriteria cũ
				if err := db.MasterDB.Model(&models.AssessmentSubcriterion{}).
					Where("criterion_id = ? AND deleted_at IS NULL", criterionID).
					Updates(map[string]interface{}{
						"deleted_at": now,
						"deleted_by": userID,
						"updated_at": now,
						"updated_by": userID,
					}).Error; err != nil {
					return nil, err
				}
			}
		} else {
			// Create new criterion
			hasSubcriteria := len(criterionInput.Subcriteria) > 0
			maxScore := criterionInput.MaxScore
			if hasSubcriteria {
				maxScore = s.sumSubcriteriaInputs(criterionInput.Subcriteria)
			}

			criterion := &models.AssessmentCriterion{
				SubjectID:      existingGroup.SubjectID,
				Name:           criterionInput.Name,
				Description:    criterionInput.Description,
				HasSubcriteria: hasSubcriteria,
				MaxScore:       maxScore,
				CreatedBy:      userID,
				CreatedAt:      now,
				UpdatedBy:      userID,
				UpdatedAt:      now,
			}

			if err := s.criterionRepo.Create(criterion); err != nil {
				return nil, err
			}

			criterionID = criterion.ID

			// Tạo subcriteria nếu có
			if hasSubcriteria {
				for _, subInput := range criterionInput.Subcriteria {
					subcriterion := &models.AssessmentSubcriterion{
						CriterionID:  criterionID,
						Name:         subInput.Name,
						Description:  subInput.Description,
						MaxScore:     subInput.MaxScore,
						CreatedBy:    userID,
						CreatedAt:    now,
						UpdatedBy:    userID,
						UpdatedAt:    now,
					}

					if err := s.subcriterionRepo.Create(subcriterion); err != nil {
						return nil, err
					}
				}
			}
		}

		newCriteriaIDs = append(newCriteriaIDs, criterionID)
	}

	// Xử lý sync criteria với group: add/restore/delete
	currentCriteriaMap := make(map[int64]bool)
	for _, id := range currentCriteriaIDs {
		currentCriteriaMap[id] = true
	}

	newCriteriaMap := make(map[int64]bool)
	for _, id := range newCriteriaIDs {
		newCriteriaMap[id] = true
	}

	// Soft delete criteria không còn trong request khỏi group
	var criteriaToDelete []int64
	for _, id := range currentCriteriaIDs {
		if !newCriteriaMap[id] {
			criteriaToDelete = append(criteriaToDelete, id)
		}
	}

	if len(criteriaToDelete) > 0 {
		if err := db.MasterDB.Model(&models.AssessmentCriteriaGroupRefCriterion{}).
			Where("assessment_criteria_group_id = ? AND assessment_criteria_id IN ? AND deleted_at IS NULL", int64(id), criteriaToDelete).
			Updates(map[string]interface{}{
				"deleted_at": now,
				"deleted_by": userID,
				"updated_at": now,
				"updated_by": userID,
			}).Error; err != nil {
			return nil, err
		}
	}

	// Add hoặc restore criteria mới vào group
	for _, criterionID := range newCriteriaIDs {
		if criterionID == 0 {
			continue
		}

		var existingRef models.AssessmentCriteriaGroupRefCriterion
		err := db.MasterDB.Unscoped().
			Where("assessment_criteria_group_id = ? AND assessment_criteria_id = ?", int64(id), criterionID).
			First(&existingRef).Error

		if err == nil && existingRef.DeletedAt.Valid {
			// Exists but soft-deleted, restore
			if err := db.MasterDB.Unscoped().
				Model(&existingRef).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"deleted_by": 0,
					"updated_by": userID,
					"updated_at": now,
				}).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			// Not found, create new
			ref := &models.AssessmentCriteriaGroupRefCriterion{
				AssessmentCriteriaGroupID: int64(id),
				AssessmentCriteriaID:      criterionID,
				CreatedBy:                 userID,
				CreatedAt:                 now,
				UpdatedBy:                 userID,
				UpdatedAt:                 now,
			}
			if err := db.MasterDB.Create(ref).Error; err != nil {
				if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
					return nil, err
				}
			}
		}
	}

	// Lấy lại group để trả về
	updatedGroup, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}

	return s.resource.FormatGroup(updatedGroup), nil
}

func (s *assessmentCriteriaGroupService) sumSubcriteriaInputs(items []*prot.AssessmentSubcriterionInput) float32 {
	var total float32
	for _, item := range items {
		total += item.MaxScore
	}
	return total
}

