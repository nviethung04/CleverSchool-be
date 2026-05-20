package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type HolidayService interface {
	GetAll(c *gin.Context) ([]models.Holiday, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Holiday, error)
	Create(c *gin.Context, req *prot.HolidayRequest) (*models.Holiday, error)
	Update(c *gin.Context, req *prot.HolidayRequest) (*models.Holiday, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Holiday, error)
}

type holidayService struct {
	repo repositories.HolidayRepository
}

func NewHolidayService(repo repositories.HolidayRepository) HolidayService {
	return &holidayService{repo: repo}
}

func (s *holidayService) GetAll(c *gin.Context) ([]models.Holiday, int64, error) {
	allowedFilters := []string{"status", "semester_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	holidays, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return holidays, rows, nil
}

func (s *holidayService) GetByID(c *gin.Context, id int) (*prot.Holiday, error) {
	holiday, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	holidayResource := resources.NewHolidayResource()
	formattedHoliday := holidayResource.FormatHoliday(holiday)

	return formattedHoliday, nil
}

func (s *holidayService) Create(c *gin.Context, req *prot.HolidayRequest) (*models.Holiday, error) {
	if req.WeekId != 0 {
		weekRepo := repositories.NewWeekRepository()
		week, err := weekRepo.FindByID(int(req.WeekId))
		if err != nil {
			return nil, err
		}

		req.StartDate = week.StartDate.Format("2006-01-02")
		req.EndDate = week.EndDate.Format("2006-01-02")

		req.Type = "week"
	}

	holidayResource := resources.NewHolidayResource()
	holiday := holidayResource.FormatModelHoliday(req)

	s.repo.SetContext(c)

	err := s.repo.Create(holiday)
	if err != nil {
		return nil, err
	}

	id := int(holiday.ID)
	newHoliday, _ := s.repo.FindNewByID(id)

	return newHoliday, nil
}

func (s *holidayService) Update(c *gin.Context, req *prot.HolidayRequest) (*models.Holiday, error) {
	if req.WeekId != 0 {
		weekRepo := repositories.NewWeekRepository()
		week, err := weekRepo.FindByID(int(req.WeekId))
		if err != nil {
			return nil, err
		}

		req.StartDate = week.StartDate.Format("2006-01-02")
		req.EndDate = week.EndDate.Format("2006-01-02")

		req.Type = "week"
	}

	holidayResource := resources.NewHolidayResource()
	holiday := holidayResource.FormatModelHoliday(req)

	s.repo.SetContext(c)

	err := s.repo.Update(holiday)
	if err != nil {
		return nil, err
	}

	id := int(holiday.ID)
	updatedHoliday, _ := s.repo.FindNewByID(id)

	return updatedHoliday, nil
}

func (s *holidayService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *holidayService) Restore(c *gin.Context, id int) (*models.Holiday, error) {
	s.repo.SetContext(c)
	holiday, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return holiday, nil
}

