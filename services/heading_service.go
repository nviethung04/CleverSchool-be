package services

import (
	"be-lms/config"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type HeadingService interface {
	GetAll(c *gin.Context) ([]models.Heading, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Heading, error)
	Create(c *gin.Context, req *prot.HeadingRequest) (*models.Heading, error)
	Update(c *gin.Context, req *prot.HeadingRequest) (*models.Heading, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Heading, error)
}

type headingService struct {
	repo repositories.HeadingRepository
}

func NewHeadingService(repo repositories.HeadingRepository) HeadingService {
	return &headingService{repo: repo}
}

func (s *headingService) GetAll(c *gin.Context) ([]models.Heading, int64, error) {
	allowedFilters := []string{"chapter_id"}
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
		"Lessons",
	})

	headings, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return headings, rows, nil
}

func (s *headingService) GetByID(c *gin.Context, id int) (*prot.Heading, error) {
	s.repo.SetPreload([]string{
		"Lessons",
	})
	heading, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	headingResource := resources.NewHeadingResource()
	formattedHeading := headingResource.FormatHeading(heading)

	return formattedHeading, nil
}

func (s *headingService) Create(c *gin.Context, req *prot.HeadingRequest) (*models.Heading, error) {
	headingResource := resources.NewHeadingResource()
	heading := headingResource.FormatModelHeading(req)

	if err := s.Validate(heading); err != nil {
		return nil, err
	}

	s.repo.SetContext(c)

	err := s.repo.Create(heading)
	if err != nil {
		return nil, err
	}

	id := int(heading.ID)
	err = s.StoreLessons(c, int64(id), req)

	if err != nil {
		config.Log.Error(err)
	}

	s.repo.SetPreload([]string{
		"Lessons",
	})
	newHeading, _ := s.repo.FindNewByID(id)

	return newHeading, nil
}

func (s *headingService) Update(c *gin.Context, req *prot.HeadingRequest) (*models.Heading, error) {
	headingResource := resources.NewHeadingResource()
	heading := headingResource.FormatModelHeading(req)

	if err := s.Validate(heading); err != nil {
		return nil, err
	}

	s.repo.SetContext(c)

	err := s.repo.Update(heading)
	if err != nil {
		return nil, err
	}

	id := int(heading.ID)
	err = s.StoreLessons(c, int64(id), req)

	if err != nil {
		config.Log.Error(err)
	}

	s.repo.SetPreload([]string{
		"Lessons",
	})
	updatedHeading, _ := s.repo.FindNewByID(id)

	return updatedHeading, nil
}

func (s *headingService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *headingService) Restore(c *gin.Context, id int) (*models.Heading, error) {
	s.repo.SetContext(c)
	heading, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return heading, nil
}

func (s *headingService) StoreLessons(c *gin.Context, id int64, req *prot.HeadingRequest) error {
	if req.Time == models.HeadingMiniTime {
		return s.repo.UpdateOrCreateLesson(id, req)
	}

	return nil
}

func (s *headingService) Validate(heading *models.Heading) error {
	mini := models.HeadingMiniTime
	value := int(heading.Time)

	if value < mini {
		return fmt.Errorf(i18n.Localize("messages.time_must_be_greater_than_or_equal", map[string]interface{}{"Value": mini}))
	}

	if value%mini != 0 {
		return fmt.Errorf(i18n.Localize("messages.time_must_be_divisible", map[string]interface{}{"Value": mini}))
	}

	isValid := s.repo.ValidTime(heading)

	if !isValid {
		return errors.New(i18n.Localize("messages.heading_time_overlap"))
	}

	return nil
}
