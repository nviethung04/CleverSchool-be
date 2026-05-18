package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

// QuestionAttributeService defines the interface for question attribute operations
type QuestionAttributeService interface {
	GetAll(c *gin.Context) ([]models.QuestionAttribute, int64, error)
	GetByID(c *gin.Context, id int) (*prot.QuestionAttribute, error)
	Create(c *gin.Context, req *prot.QuestionAttributeRequest) (*models.QuestionAttribute, error)
	Update(c *gin.Context, req *prot.QuestionAttributeRequest) (*models.QuestionAttribute, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.QuestionAttribute, error)
	GetParents(c *gin.Context) ([]models.QuestionAttribute, error)
}

type questionAttributeService struct {
	repo repositories.QuestionAttributeRepository
}

// NewQuestionAttributeService creates a new instance of QuestionAttributeService
func NewQuestionAttributeService(repo repositories.QuestionAttributeRepository) QuestionAttributeService {
	return &questionAttributeService{
		repo: repo,
	}
}

func (s *questionAttributeService) GetAll(c *gin.Context) ([]models.QuestionAttribute, int64, error) {
	allowedFilters := []string{"type", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	questionAttributes, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return questionAttributes, rows, nil
}

func (s *questionAttributeService) GetByID(c *gin.Context, id int) (*prot.QuestionAttribute, error) {
	attr, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	resource := resources.NewQuestionAttributeResource()
	resp := resource.FormatQuestionAttribute(attr)

	return resp, nil
}

func (s *questionAttributeService) Create(c *gin.Context, req *prot.QuestionAttributeRequest) (*models.QuestionAttribute, error) {
	questionAttributeResource := resources.NewQuestionAttributeResource()
	questionAttribute := questionAttributeResource.FormatModelQuestionAttribute(req)

	s.repo.SetContext(c)

	err := s.repo.Create(questionAttribute)
	if err != nil {
		return nil, err
	}

	id := int(questionAttribute.ID)
	newQuestionAttribute, _ := s.repo.FindNewByID(id)

	return newQuestionAttribute, nil
}

func (s *questionAttributeService) Update(c *gin.Context, req *prot.QuestionAttributeRequest) (*models.QuestionAttribute, error) {
	questionAttributeResource := resources.NewQuestionAttributeResource()
	questionAttribute := questionAttributeResource.FormatModelQuestionAttribute(req)

	s.repo.SetContext(c)

	err := s.repo.Update(questionAttribute)
	if err != nil {
		return nil, err
	}

	id := int(questionAttribute.ID)
	updatedQuestionAttribute, _ := s.repo.FindNewByID(id)

	return updatedQuestionAttribute, nil
}

func (s *questionAttributeService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *questionAttributeService) Restore(c *gin.Context, id int) (*models.QuestionAttribute, error) {
	s.repo.SetContext(c)
	questionAttribute, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return questionAttribute, nil
}

func (s *questionAttributeService) GetParents(c *gin.Context) ([]models.QuestionAttribute, error) {
	allowedFilters := []string{"subject_id"}
	filter, _, _, _, _, err := utils.ParsePaginationParams(c, allowedFilters)
	s.repo.SetFilter(filter)

	if err != nil {
		return nil, err
	}
	attributes, err := s.repo.GetParents()
	return attributes, err
}
