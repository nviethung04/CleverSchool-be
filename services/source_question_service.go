package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type SourceQuestionService interface {
	GetAll(c *gin.Context) ([]models.SourceQuestion, int64, error)
	GetByID(c *gin.Context, id int) (*models.SourceQuestion, error)
	Create(c *gin.Context, req *prot.SourceQuestionRequest) (*models.SourceQuestion, error)
	Update(c *gin.Context, req *prot.SourceQuestionRequest) (*models.SourceQuestion, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.SourceQuestion, error)
}

type sourceQuestionService struct {
	repo repositories.SourceQuestionRepository
}

func NewSourceQuestionService(repo repositories.SourceQuestionRepository) SourceQuestionService {
	return &sourceQuestionService{repo}
}

func (s *sourceQuestionService) GetAll(c *gin.Context) ([]models.SourceQuestion, int64, error) {
	allowedFilters := []string{"status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"title", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	sourceQuestions, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return sourceQuestions, rows, nil
}

func (s *sourceQuestionService) GetByID(c *gin.Context, id int) (*models.SourceQuestion, error) {
	return s.repo.FindByID(id)
}

func (s *sourceQuestionService) Create(c *gin.Context, req *prot.SourceQuestionRequest) (*models.SourceQuestion, error) {
	sourceQuestionResource := resources.NewSourceQuestionResource()
	sourceQuestion := sourceQuestionResource.FormatModelSourceQuestion(req)

	s.repo.SetContext(c)

	err := s.repo.Create(sourceQuestion)
	if err != nil {
		return nil, err
	}

	id := int(sourceQuestion.ID)
	newSourceQuestion, _ := s.repo.FindNewByID(id)

	return newSourceQuestion, nil
}

func (s *sourceQuestionService) Update(c *gin.Context, req *prot.SourceQuestionRequest) (*models.SourceQuestion, error) {
	sourceQuestionResource := resources.NewSourceQuestionResource()
	sourceQuestion := sourceQuestionResource.FormatModelSourceQuestion(req)

	s.repo.SetContext(c)

	err := s.repo.Update(sourceQuestion)
	if err != nil {
		return nil, err
	}

	id := int(sourceQuestion.ID)
	updateSourceQuestion, _ := s.repo.FindNewByID(id)

	return updateSourceQuestion, nil
}

func (s *sourceQuestionService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *sourceQuestionService) Restore(c *gin.Context, id int) (*models.SourceQuestion, error) {
	s.repo.SetContext(c)
	rource, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return rource, nil
}
