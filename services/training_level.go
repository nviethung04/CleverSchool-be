package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type TrainingLevelService interface {
	GetAll(c *gin.Context) ([]models.TrainingLevel, int64, error)
	GetByID(c *gin.Context, id int) (*prot.TrainingLevel, error)
	Create(c *gin.Context, req *prot.TrainingLevelRequest) (*models.TrainingLevel, error)
	Update(c *gin.Context, req *prot.TrainingLevelRequest) (*models.TrainingLevel, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.TrainingLevel, error)
}

type trainingLevelService struct {
	repo repositories.TrainingLevelRepository
}

func NewTrainingLevelService(repo repositories.TrainingLevelRepository) TrainingLevelService {
	return &trainingLevelService{repo: repo}
}

func (s *trainingLevelService) GetAll(c *gin.Context) ([]models.TrainingLevel, int64, error) {
	allowedFilters := []string{}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	trainingLevels, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return trainingLevels, rows, nil
}

func (s *trainingLevelService) GetByID(c *gin.Context, id int) (*prot.TrainingLevel, error) {
	trainingLevel, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	trainingLevelResource := resources.NewTrainingLevelResource()
	formattedTrainingLevel := trainingLevelResource.FormatTrainingLevel(trainingLevel)

	return formattedTrainingLevel, nil
}

func (s *trainingLevelService) Create(c *gin.Context, req *prot.TrainingLevelRequest) (*models.TrainingLevel, error) {
	trainingLevelResource := resources.NewTrainingLevelResource()
	trainingLevel := trainingLevelResource.FormatModelTrainingLevel(req)

	s.repo.SetContext(c)

	err := s.repo.Create(trainingLevel)
	if err != nil {
		return nil, err
	}

	id := int(trainingLevel.ID)
	newTrainingLevel, _ := s.repo.FindNewByID(id)

	return newTrainingLevel, nil
}

func (s *trainingLevelService) Update(c *gin.Context, req *prot.TrainingLevelRequest) (*models.TrainingLevel, error) {
	trainingLevelResource := resources.NewTrainingLevelResource()
	trainingLevel := trainingLevelResource.FormatModelTrainingLevel(req)

	s.repo.SetContext(c)

	err := s.repo.Update(trainingLevel)
	if err != nil {
		return nil, err
	}

	id := int(trainingLevel.ID)
	updatedTrainingLevel, _ := s.repo.FindNewByID(id)

	return updatedTrainingLevel, nil
}

func (s *trainingLevelService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *trainingLevelService) Restore(c *gin.Context, id int) (*models.TrainingLevel, error) {
	s.repo.SetContext(c)
	trainingLevel, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return trainingLevel, nil
}
