package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DepartmentService interface {
	GetAll(c *gin.Context) ([]models.Department, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Department, error)
	Create(c *gin.Context, req *prot.Department) (*models.Department, error)
	Update(c *gin.Context, req *prot.Department) (*models.Department, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Department, error)
}

type departmentService struct {
	repo repositories.DepartmentRepository
}

func NewDepartmentService(repo repositories.DepartmentRepository) DepartmentService {
	return &departmentService{repo}
}

func (s *departmentService) GetAll(c *gin.Context) ([]models.Department, int64, error) {
	allowedFilters := []string{"status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	departmentes, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return departmentes, rows, nil
}

func (s *departmentService) GetByID(c *gin.Context, id int) (*prot.Department, error) {
	department, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	departmentResource := resources.NewDepartmentResource()
	formattedDepartment := departmentResource.FormatDepartment(department)

	return formattedDepartment, nil
}

func (s *departmentService) Create(c *gin.Context, req *prot.Department) (*models.Department, error) {
	departmentResource := resources.NewDepartmentResource()
	department := departmentResource.FormatModelDepartment(req)

	s.repo.SetContext(c)

	err := s.repo.Create(department)
	if err != nil {
		return nil, err
	}

	id := int(department.ID)
	newDepartment, _ := s.repo.FindNewByID(id)

	return newDepartment, nil
}

func (s *departmentService) Update(c *gin.Context, req *prot.Department) (*models.Department, error) {
	departmentResource := resources.NewDepartmentResource()
	department := departmentResource.FormatModelDepartment(req)

	s.repo.SetContext(c)

	err := s.repo.Update(department)
	if err != nil {
		return nil, err
	}

	id := int(department.ID)
	updateDepartment, _ := s.repo.FindNewByID(id)

	return updateDepartment, nil
}

func (s *departmentService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *departmentService) Restore(c *gin.Context, id int) (*models.Department, error) {
	s.repo.SetContext(c)
	department, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return department, nil
}
