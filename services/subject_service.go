package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type SubjectService interface {
	GetAll(c *gin.Context) ([]models.Subject, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Subject, error)
	Create(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error)
	Update(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Subject, error)
}

type subjectService struct {
	repo repositories.SubjectRepository
}

func NewSubjectService(repo repositories.SubjectRepository) SubjectService {
	return &subjectService{repo}
}

func (s *subjectService) GetAll(c *gin.Context) ([]models.Subject, int64, error) {
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

	subjects, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return subjects, rows, nil
}

func (s *subjectService) GetByID(c *gin.Context, id int) (*prot.Subject, error) {
	subject, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	subjectResource := resources.NewSubjectResource()
	formattedSubject := subjectResource.FormatSubject(subject)

	return formattedSubject, nil
}

func (s *subjectService) Create(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error) {
	subjectResource := resources.NewSubjectResource()
	subject := subjectResource.FormatModelSubject(req)

	s.repo.SetContext(c)

	err := s.repo.Create(subject)
	if err != nil {
		return nil, err
	}

	id := int(subject.ID)
	newSubject, _ := s.repo.FindNewByID(id)

	return newSubject, nil
}

func (s *subjectService) Update(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error) {
	subjectResource := resources.NewSubjectResource()
	subject := subjectResource.FormatModelSubject(req)

	s.repo.SetContext(c)

	err := s.repo.Update(subject)
	if err != nil {
		return nil, err
	}

	id := int(subject.ID)
	updateSubject, _ := s.repo.FindNewByID(id)

	return updateSubject, nil
}

func (s *subjectService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *subjectService) Restore(c *gin.Context, id int) (*models.Subject, error) {
	s.repo.SetContext(c)
	subject, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return subject, nil
}
