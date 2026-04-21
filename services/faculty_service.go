package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type FacultyService interface {
	GetAll(c *gin.Context) ([]models.Faculty, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Faculty, error)
	Create(c *gin.Context, req *prot.FacultyRequest) (*models.Faculty, error)
	Update(c *gin.Context, req *prot.FacultyRequest) (*models.Faculty, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Faculty, error)
	Export(c *gin.Context) (string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
}

type facultyService struct {
	repo repositories.FacultyRepository
}

func NewFacultyService(repo repositories.FacultyRepository) FacultyService {
	return &facultyService{repo: repo}
}

func (s *facultyService) GetAll(c *gin.Context) ([]models.Faculty, int64, error) {
	allowedFilters := []string{"code", "status", "school_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{"School"})

	faculties, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return faculties, rows, nil
}

func (s *facultyService) GetByID(c *gin.Context, id int) (*prot.Faculty, error) {
	s.repo.SetPreload([]string{"School"})
	s.repo.SetContext(c)
	faculty, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	facultyResource := resources.NewFacultyResource()
	formattedFaculty := facultyResource.FormatFaculty(faculty)

	return formattedFaculty, nil
}

func (s *facultyService) Create(c *gin.Context, req *prot.FacultyRequest) (*models.Faculty, error) {
	facultyResource := resources.NewFacultyResource()
	faculty := facultyResource.FormatModelFaculty(req)

	s.repo.SetContext(c)

	err := s.repo.Create(faculty)
	if err != nil {
		return nil, err
	}

	id := int(faculty.ID)
	s.repo.SetPreload([]string{"School"})
	newFaculty, _ := s.repo.FindNewByID(id)

	return newFaculty, nil
}

func (s *facultyService) Update(c *gin.Context, req *prot.FacultyRequest) (*models.Faculty, error) {
	facultyResource := resources.NewFacultyResource()
	faculty := facultyResource.FormatModelFaculty(req)

	s.repo.SetContext(c)

	err := s.repo.Update(faculty)
	if err != nil {
		return nil, err
	}

	id := int(faculty.ID)
	s.repo.SetPreload([]string{"School"})
	updatedFaculty, _ := s.repo.FindNewByID(id)

	return updatedFaculty, nil
}

func (s *facultyService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *facultyService) Restore(c *gin.Context, id int) (*models.Faculty, error) {
	s.repo.SetContext(c)
	faculty, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return faculty, nil
}
