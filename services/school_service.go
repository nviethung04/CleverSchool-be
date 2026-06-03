package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type SchoolService interface {
	GetAll(c *gin.Context) ([]models.School, int64, error)
	GetByID(c *gin.Context, id int) (*prot.School, error)
	Create(c *gin.Context, req *prot.SchoolRequest) (*models.School, error)
	Update(c *gin.Context, req *prot.SchoolRequest) (*models.School, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.School, error)
	Export(c *gin.Context) (string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
	GetSchoolStudents(c *gin.Context, schoolId int64) ([]models.User, error)
}

type schoolService struct {
	repo repositories.SchoolRepository
}

func NewSchoolService(repo repositories.SchoolRepository) SchoolService {
	return &schoolService{repo: repo}
}

func (s *schoolService) GetAll(c *gin.Context) ([]models.School, int64, error) {
	var req requests.GetSchoolRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	schools, total, err := s.repo.GetAllSchool(&req, c)
	if err != nil {
		return nil, 0, err
	}

	return schools, total, err
}

func (s *schoolService) GetByID(c *gin.Context, id int) (*prot.School, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Ward",
		"Ward.Province",
	})

	school, err := s.repo.FindByID(int(id))

	if err != nil {
		return nil, err
	}

	schoolResource := resources.NewSchoolResource()
	formatSchool := schoolResource.FormatSchool(school)

	return formatSchool, nil
}
func (s *schoolService) Create(c *gin.Context, req *prot.SchoolRequest) (*models.School, error) {
	schoolResource := resources.NewSchoolResource()
	school := schoolResource.FormatModelSchool(req)

	s.repo.SetContext(c)

	if err := s.repo.Create(school); err != nil {
		return nil, err
	}

	id := int(school.ID)

	s.repo.SetPreload([]string{
		"Ward",
		"Ward.Province",
	})

	newSchool, _ := s.repo.FindNewByID(id)

	return newSchool, nil
}

func (s *schoolService) Update(c *gin.Context, req *prot.SchoolRequest) (*models.School, error) {
	schoolResource := resources.NewSchoolResource()
	schoolUpdate := schoolResource.FormatModelSchool(req)

	s.repo.SetContext(c)

	if err := s.repo.Update(schoolUpdate); err != nil {
		return nil, err
	}

	id := int(schoolUpdate.ID)

	classRepo := repositories.NewClassRepository()
	classRepo.UpdateSchoolStats(schoolUpdate.ID)

	s.repo.SetPreload([]string{
		"Ward",
		"Ward.Province",
	})

	newSchool, _ := s.repo.FindNewByID(id)

	return newSchool, nil
}

func (s *schoolService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(int(id))
}

func (s *schoolService) Restore(c *gin.Context, id int) (*models.School, error) {
	s.repo.SetContext(c)
	school, err := s.repo.Restore(int(id))
	if err != nil {
		return nil, err
	}
	return school, nil
}

func (s *schoolService) GetSchoolStudents(c *gin.Context, schoolId int64) ([]models.User, error) {
	s.repo.SetContext(c)
	students, err := s.repo.GetSchoolStudents(schoolId)
	if err != nil {
		return nil, err
	}
	return students, nil
}
