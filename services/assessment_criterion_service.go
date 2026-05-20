package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentCriterionService interface {
	GetAll(c *gin.Context) ([]models.AssessmentCriterion, int64, error)
	GetByID(c *gin.Context, id int) (*prot.AssessmentCriterion, error)
	Create(c *gin.Context, req *prot.AssessmentCriterionRequest) (*models.AssessmentCriterion, error)
	Update(c *gin.Context, req *prot.AssessmentCriterionRequest) (*models.AssessmentCriterion, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.AssessmentCriterion, error)
	CreateBulk(c *gin.Context, req *prot.CreateAssessmentCriteriaRequest) error
}

type assessmentCriterionService struct {
	repo             repositories.AssessmentCriterionRepository
	subcriterionRepo repositories.AssessmentSubcriterionRepository
	resource         resources.AssessmentCriterionResource
}

func NewAssessmentCriterionService(repo repositories.AssessmentCriterionRepository) AssessmentCriterionService {
	return &assessmentCriterionService{
		repo:             repo,
		subcriterionRepo: repositories.NewAssessmentSubcriterionRepository(),
		resource:         resources.NewAssessmentCriterionResource(),
	}
}

func (s *assessmentCriterionService) GetAll(c *gin.Context) ([]models.AssessmentCriterion, int64, error) {
	var req requests.GetAssessmentCriterionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	return s.repo.GetAllWithPaging(&req, c)
}

func (s *assessmentCriterionService) GetByID(c *gin.Context, id int) (*prot.AssessmentCriterion, error) {
	item, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}
	return s.resource.FormatAssessmentCriterion(item), nil
}

func (s *assessmentCriterionService) Create(c *gin.Context, req *prot.AssessmentCriterionRequest) (*models.AssessmentCriterion, error) {
	entity := s.resource.FormatModelAssessmentCriterion(req)
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

	return entity, nil
}

func (s *assessmentCriterionService) Update(c *gin.Context, req *prot.AssessmentCriterionRequest) (*models.AssessmentCriterion, error) {
	entity := s.resource.FormatModelAssessmentCriterion(req)
	if entity == nil {
		return nil, nil
	}
	entity.UpdatedBy = int64(utils.GetCurrentUserId(c))
	entity.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *assessmentCriterionService) Delete(c *gin.Context, id int) error {
	return s.repo.Delete(int64(id), int64(utils.GetCurrentUserId(c)))
}

func (s *assessmentCriterionService) Restore(c *gin.Context, id int) (*models.AssessmentCriterion, error) {
	baseRepo := base.NewBaseRepository[*models.AssessmentCriterion]()
	baseRepo.SetContext(c)
	item, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *item, nil
}

func (s *assessmentCriterionService) CreateBulk(c *gin.Context, req *prot.CreateAssessmentCriteriaRequest) error {
	userID := int64(utils.GetCurrentUserId(c))

	if req.SubjectId == 0 {
		return fmt.Errorf("subject_id is required")
	}

	now := time.Now().UTC()

	for _, criterionItem := range req.Criteria {
		hasSubcriteria := len(criterionItem.Subcriteria) > 0
		maxScore := criterionItem.MaxScore
		if hasSubcriteria {
			maxScore = s.sumSubcriteriaScores(criterionItem.Subcriteria)
		}
		criterion := &models.AssessmentCriterion{
			SubjectID:      req.SubjectId,
			Name:           criterionItem.Name,
			Description:    criterionItem.Description,
			HasSubcriteria: hasSubcriteria,
			MaxScore:       maxScore,
			CreatedBy:      userID,
			CreatedAt:      now,
			UpdatedBy:      userID,
			UpdatedAt:      now,
		}

		if err := s.repo.Create(criterion); err != nil {
			return err
		}

		// Create subcriteria if exists
		if len(criterionItem.Subcriteria) > 0 {
			for _, subcriterionItem := range criterionItem.Subcriteria {
				subcriterion := &models.AssessmentSubcriterion{
					CriterionID: criterion.ID,
					Name:        subcriterionItem.Name,
					Description: subcriterionItem.Description,
					MaxScore:    subcriterionItem.MaxScore,
					CreatedBy:   userID,
					CreatedAt:   now,
					UpdatedBy:   userID,
					UpdatedAt:   now,
				}

				if err := s.subcriterionRepo.Create(subcriterion); err != nil {
					return err
				}
			}
		}

		// If assessment_id is provided, create ref
		if req.AssessmentId > 0 {
			ref := &models.AssessmentRefCriterion{
				AssessmentID:          req.AssessmentId,
				AssessmentCriterionID: criterion.ID,
				CreatedBy:             userID,
				CreatedAt:             now,
				UpdatedBy:             userID,
				UpdatedAt:             now,
			}

			if err := s.repo.CreateRefCriterion(ref); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *assessmentCriterionService) sumSubcriteriaScores(items []*prot.AssessmentSubcriterionItem) float32 {
	var total float32
	for _, item := range items {
		total += item.MaxScore
	}
	return total
}

