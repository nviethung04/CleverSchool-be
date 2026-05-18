package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type StudyShiftService interface {
	GetAll(c *gin.Context) ([]models.StudyShift, int64, error)
	GetByID(c *gin.Context, id int) (*prot.StudyShift, error)
	Create(c *gin.Context, req *prot.StudyShiftRequest) (*models.StudyShift, error)
	Update(c *gin.Context, req *prot.StudyShiftRequest) (*models.StudyShift, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.StudyShift, error)
}

type studyShiftService struct {
	repo repositories.StudyShiftRepository
}

func NewStudyShiftService(repo repositories.StudyShiftRepository) StudyShiftService {
	return &studyShiftService{repo: repo}
}

func (s *studyShiftService) GetAll(c *gin.Context) ([]models.StudyShift, int64, error) {
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

	studyShifts, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return studyShifts, rows, nil
}

func (s *studyShiftService) GetByID(c *gin.Context, id int) (*prot.StudyShift, error) {
	s.repo.SetPreload([]string{})
	studyShift, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	studyShiftResource := resources.NewStudyShiftResource()
	formattedStudyShift := studyShiftResource.FormatStudyShift(studyShift)

	return formattedStudyShift, nil
}

func (s *studyShiftService) Create(c *gin.Context, req *prot.StudyShiftRequest) (*models.StudyShift, error) {
	studyShiftResource := resources.NewStudyShiftResource()
	studyShift := studyShiftResource.FormatModelStudyShift(req)

	s.repo.SetContext(c)

	if err := s.repo.Create(studyShift); err != nil {
		return nil, err
	}

	id := int(studyShift.ID)

	s.repo.SetPreload([]string{})

	newStudyShift, _ := s.repo.FindNewByID(id)

	return newStudyShift, nil
}

func (s *studyShiftService) Update(c *gin.Context, req *prot.StudyShiftRequest) (*models.StudyShift, error) {
	studyShiftResource := resources.NewStudyShiftResource()
	studyShift := studyShiftResource.FormatModelStudyShift(req)

	s.repo.SetContext(c)

	if err := s.repo.Update(studyShift); err != nil {
		return nil, err
	}

	id := int(studyShift.ID)

	s.repo.SetPreload([]string{})

	updatedStudyShift, _ := s.repo.FindNewByID(id)

	return updatedStudyShift, nil
}

func (s *studyShiftService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *studyShiftService) Restore(c *gin.Context, id int) (*models.StudyShift, error) {
	s.repo.SetContext(c)
	studyShift, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return studyShift, nil
}
