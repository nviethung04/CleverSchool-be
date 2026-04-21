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

type SubjectService interface {
	GetAll(c *gin.Context) ([]models.Subject, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Subject, error)
	Create(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error)
	Update(c *gin.Context, req *prot.SubjectRequest) (*models.Subject, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Subject, error)
	Export(c *gin.Context) (string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
}

type subjectService struct {
	repo repositories.SubjectRepository
}

func NewSubjectService(repo repositories.SubjectRepository) SubjectService {
	return &subjectService{repo}
}

func (s *subjectService) GetAll(c *gin.Context) ([]models.Subject, int64, error) {
	allowedFilters := []string{"status", "faculty_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Faculty",
		"TrainingLevels",
	})
	s.repo.SetContext(c)

	subjects, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return subjects, rows, nil
}

func (s *subjectService) GetByID(c *gin.Context, id int) (*prot.Subject, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Faculty",
		"TrainingLevels",
	})
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

	s.StoreTrainingLevels(c, subject.ID, req.TrainingLevels)

	id := int(subject.ID)
	s.repo.SetPreload([]string{
		"Faculty",
		"TrainingLevels",
	})
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

	s.StoreTrainingLevels(c, subject.ID, req.TrainingLevels)

	id := int(subject.ID)
	s.repo.SetPreload([]string{
		"Faculty",
		"TrainingLevels",
	})
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

func (s *subjectService) StoreTrainingLevels(c *gin.Context, id int64, trainingLevels []*prot.TrainingLevel) error {
	var trainingLevelIds []int64

	for _, trainingLevel := range trainingLevels {
		trainingLevelIds = append(trainingLevelIds, trainingLevel.Id)
	}

	s.repo.UpdateTrainingLevels(id, trainingLevelIds)

	return nil
}
