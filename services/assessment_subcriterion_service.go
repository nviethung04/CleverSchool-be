package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/repositories/base"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentSubcriterionService interface {
	GetAll(c *gin.Context) ([]models.AssessmentSubcriterion, int64, error)
	GetByID(c *gin.Context, id int) (*prot.AssessmentSubcriterion, error)
	Create(c *gin.Context, req *prot.AssessmentSubcriterionRequest) (*models.AssessmentSubcriterion, error)
	Update(c *gin.Context, req *prot.AssessmentSubcriterionRequest) (*models.AssessmentSubcriterion, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.AssessmentSubcriterion, error)
}

type assessmentSubcriterionService struct {
	repo     repositories.AssessmentSubcriterionRepository
	resource resources.AssessmentSubcriterionResource
}

func NewAssessmentSubcriterionService(repo repositories.AssessmentSubcriterionRepository) AssessmentSubcriterionService {
	return &assessmentSubcriterionService{
		repo:     repo,
		resource: resources.NewAssessmentSubcriterionResource(),
	}
}

func (s *assessmentSubcriterionService) GetAll(c *gin.Context) ([]models.AssessmentSubcriterion, int64, error) {
	var req requests.GetAssessmentSubcriterionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	return s.repo.GetAllWithPaging(&req, c)
}

func (s *assessmentSubcriterionService) GetByID(c *gin.Context, id int) (*prot.AssessmentSubcriterion, error) {
	item, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}
	return s.resource.FormatAssessmentSubcriterion(item), nil
}

func (s *assessmentSubcriterionService) Create(c *gin.Context, req *prot.AssessmentSubcriterionRequest) (*models.AssessmentSubcriterion, error) {
	entity := s.resource.FormatModelAssessmentSubcriterion(req)
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

func (s *assessmentSubcriterionService) Update(c *gin.Context, req *prot.AssessmentSubcriterionRequest) (*models.AssessmentSubcriterion, error) {
	entity := s.resource.FormatModelAssessmentSubcriterion(req)
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

func (s *assessmentSubcriterionService) Delete(c *gin.Context, id int) error {
	return s.repo.Delete(int64(id), int64(utils.GetCurrentUserId(c)))
}

func (s *assessmentSubcriterionService) Restore(c *gin.Context, id int) (*models.AssessmentSubcriterion, error) {
	baseRepo := base.NewBaseRepository[*models.AssessmentSubcriterion]()
	baseRepo.SetContext(c)
	item, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *item, nil
}
