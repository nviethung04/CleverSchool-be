package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type DegreeService interface {
	GetAll(c *gin.Context) ([]models.Degree, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Degree, error)
	Create(c *gin.Context, req *prot.Degree) (*models.Degree, error)
	Update(c *gin.Context, req *prot.Degree) (*models.Degree, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Degree, error)
}

type degreeService struct {
	repo repositories.DegreeRepository
}

func NewDegreeService(repo repositories.DegreeRepository) DegreeService {
	return &degreeService{repo}
}

func (s *degreeService) GetAll(c *gin.Context) ([]models.Degree, int64, error) {
	allowedFilters := []string{"user_id", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	degreees, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return degreees, rows, nil
}

func (s *degreeService) GetByID(c *gin.Context, id int) (*prot.Degree, error) {
	degree, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	degreeResource := resources.NewDegreeResource()
	formattedDegree := degreeResource.FormatDegree(degree)

	return formattedDegree, nil
}

func (s *degreeService) Create(c *gin.Context, req *prot.Degree) (*models.Degree, error) {
	degreeResource := resources.NewDegreeResource()
	degree := degreeResource.FormatModelDegree(req)

	s.repo.SetContext(c)

	err := s.repo.Create(degree)
	if err != nil {
		return nil, err
	}

	id := int(degree.ID)
	newDegree, _ := s.repo.FindNewByID(id)

	return newDegree, nil
}

func (s *degreeService) Update(c *gin.Context, req *prot.Degree) (*models.Degree, error) {
	degreeResource := resources.NewDegreeResource()
	degree := degreeResource.FormatModelDegree(req)

	s.repo.SetContext(c)

	err := s.repo.Update(degree)
	if err != nil {
		return nil, err
	}

	id := int(degree.ID)
	updateDegree, _ := s.repo.FindNewByID(id)

	return updateDegree, nil
}

func (s *degreeService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *degreeService) Restore(c *gin.Context, id int) (*models.Degree, error) {
	s.repo.SetContext(c)
	degree, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return degree, nil
}

