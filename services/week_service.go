package services

import (
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WeekService interface {
	GetAll(c *gin.Context) ([]models.Week, int64, error)
	GetByDate(date string) (*models.Week, error)
}

type weekService struct {
	repo repositories.WeekRepository
}

func NewWeekService(repo repositories.WeekRepository) WeekService {
	return &weekService{repo: repo}
}

func (s *weekService) GetAll(c *gin.Context) ([]models.Week, int64, error) {
	allowedFilters := []string{}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetSearch(keyword, []string{"week_number", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	weeks, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return weeks, rows, nil
}

func (s *weekService) GetByDate(date string) (*models.Week, error) {
	parsedDate, err := utils.ParseDate(date)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByDate(parsedDate)
}

func (s *weekService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	if courseIDStr := c.Query("course_id"); courseIDStr != "" {
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			return filter, err
		}
		courseRepo := repositories.NewCourseRepository()
		courseRepo.SetContext(c)
		course, _ := courseRepo.FindByID(int(courseID))

		if course.ID != 0 {
			if !course.StartDate.IsZero() && !course.EndDate.IsZero() {
				startStr := course.StartDate.Format("2006-01-02")
				endStr   := course.EndDate.Format("2006-01-02")

				filter["end_date"] = ">=:" + startStr
				filter["start_date"] = "<:" + endStr
			}
		}
	}

	if semesterIDStr := c.Query("semester_id"); semesterIDStr != "" {
		semesterID, err := strconv.ParseInt(semesterIDStr, 10, 64)
		if err != nil {
			return filter, err
		}
		semesterRepo := repositories.NewSemesterRepository()
		semesterRepo.SetContext(c)
		semester, _ := semesterRepo.FindByID(int(semesterID))

		if semester.ID != 0 {
			if !semester.StartDate.IsZero() && !semester.EndDate.IsZero() {
				startStr := semester.StartDate.Format("2006-01-02")
				endStr   := semester.EndDate.Format("2006-01-02")

				filter["end_date"] = ">=:" + startStr
				filter["start_date"] = "<:" + endStr
			}
		}
	}

	return filter, nil
}
