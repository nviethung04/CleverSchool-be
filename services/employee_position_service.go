package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type EmployeePositionService interface {
	GetAll(c *gin.Context) ([]models.EmployeePosition, int64, error)
	GetByID(c *gin.Context, id int) (*prot.EmployeePosition, error)
	Create(c *gin.Context, req *prot.EmployeePosition) (*models.EmployeePosition, error)
	Update(c *gin.Context, req *prot.EmployeePosition) (*models.EmployeePosition, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.EmployeePosition, error)
}

type employeePositionService struct {
	repo repositories.EmployeePositionRepository
}

func NewEmployeePositionService(repo repositories.EmployeePositionRepository) EmployeePositionService {
	return &employeePositionService{repo}
}

func (s *employeePositionService) GetAll(c *gin.Context) ([]models.EmployeePosition, int64, error) {
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

	employeePositiones, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return employeePositiones, rows, nil
}

func (s *employeePositionService) GetByID(c *gin.Context, id int) (*prot.EmployeePosition, error) {
	employeePosition, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	employeePositionResource := resources.NewEmployeePositionResource()
	formattedEmployeePosition := employeePositionResource.FormatEmployeePosition(employeePosition)

	return formattedEmployeePosition, nil
}

func (s *employeePositionService) Create(c *gin.Context, req *prot.EmployeePosition) (*models.EmployeePosition, error) {
	employeePositionResource := resources.NewEmployeePositionResource()
	employeePosition := employeePositionResource.FormatModelEmployeePosition(req)

	s.repo.SetContext(c)

	err := s.repo.Create(employeePosition)
	if err != nil {
		return nil, err
	}

	id := int(employeePosition.ID)
	newEmployeePosition, _ := s.repo.FindNewByID(id)

	return newEmployeePosition, nil
}

func (s *employeePositionService) Update(c *gin.Context, req *prot.EmployeePosition) (*models.EmployeePosition, error) {
	employeePositionResource := resources.NewEmployeePositionResource()
	employeePosition := employeePositionResource.FormatModelEmployeePosition(req)

	s.repo.SetContext(c)

	err := s.repo.Update(employeePosition)
	if err != nil {
		return nil, err
	}

	id := int(employeePosition.ID)
	updateEmployeePosition, _ := s.repo.FindNewByID(id)

	return updateEmployeePosition, nil
}

func (s *employeePositionService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *employeePositionService) Restore(c *gin.Context, id int) (*models.EmployeePosition, error) {
	s.repo.SetContext(c)
	employeePosition, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return employeePosition, nil
}
