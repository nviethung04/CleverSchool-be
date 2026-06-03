package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type GradeService interface {
	GetAll(c *gin.Context) ([]models.Grade, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Grade, error)
	Create(c *gin.Context, req *prot.GradeRequest) (*models.Grade, error)
	Update(c *gin.Context, req *prot.GradeRequest) (*models.Grade, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Grade, error)
}

type gradeService struct {
	repo repositories.GradeRepository
}

func NewGradeService(repo repositories.GradeRepository) GradeService {
	return &gradeService{repo: repo}
}

func (s *gradeService) GetAll(c *gin.Context) ([]models.Grade, int64, error) {
	allowedFilters := []string{}
	filter, page, perPage, _, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	grades, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return grades, rows, nil
}

func (s *gradeService) GetByID(c *gin.Context, id int) (*prot.Grade, error) {
	grade, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	gradeResource := resources.NewGradeResource()
	formattedGrade := gradeResource.FormatGrade(grade)

	return formattedGrade, nil
}

func (s *gradeService) Create(c *gin.Context, req *prot.GradeRequest) (*models.Grade, error) {
	gradeResource := resources.NewGradeResource()
	grade := gradeResource.FormatModelGrade(req)

	s.repo.SetContext(c)

	err := s.repo.Create(grade)
	if err != nil {
		return nil, err
	}

	id := int(grade.ID)
	newGrade, _ := s.repo.FindNewByID(id)

	return newGrade, nil
}

func (s *gradeService) Update(c *gin.Context, req *prot.GradeRequest) (*models.Grade, error) {
	gradeResource := resources.NewGradeResource()
	grade := gradeResource.FormatModelGrade(req)

	s.repo.SetContext(c)

	err := s.repo.Update(grade)
	if err != nil {
		return nil, err
	}

	id := int(grade.ID)
	updatedGrade, _ := s.repo.FindNewByID(id)

	return updatedGrade, nil
}

func (s *gradeService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *gradeService) Restore(c *gin.Context, id int) (*models.Grade, error) {
	s.repo.SetContext(c)
	grade, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return grade, nil
}
