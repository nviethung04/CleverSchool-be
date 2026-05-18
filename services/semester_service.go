package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SemesterService interface {
	GetAll(c *gin.Context) ([]models.Semester, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Semester, error)
	Create(c *gin.Context, req *prot.SemesterRequest) (*models.Semester, error)
	Update(c *gin.Context, req *prot.SemesterRequest) (*models.Semester, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Semester, error)
}

type semesterService struct {
	repo repositories.SemesterRepository
}

func NewSemesterService(repo repositories.SemesterRepository) SemesterService {
	return &semesterService{repo: repo}
}

func (s *semesterService) GetAll(c *gin.Context) ([]models.Semester, int64, error) {
	allowedFilters := []string{"status"}
	filter, page, perPage, keyword, _, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetPreload([]string{
		"Holidays",
	})

	sort := map[string]string{
		"id": "asc",
	}
	s.repo.SetSort(sort)

	semesters, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return semesters, rows, nil
}

func (s *semesterService) GetByID(c *gin.Context, id int) (*prot.Semester, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Holidays",
	})
	semester, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	semesterResource := resources.NewSemesterResource()
	formattedSemester := semesterResource.FormatSemester(semester)

	return formattedSemester, nil
}

func (s *semesterService) Create(c *gin.Context, req *prot.SemesterRequest) (*models.Semester, error) {
	semesterResource := resources.NewSemesterResource()

	if req.PreviousSemesterId != 0 {
		beginDate := s.GetEarliestStartDateFromPrevious(int(req.PreviousSemesterId), map[int]bool{})
		if beginDate != nil {
			req.BeginDate = beginDate.Format("2006-01-02")
		}
	}

	semester := semesterResource.FormatModelSemester(req)
	weekRepo := repositories.NewWeekRepository()

	if req.StartWeekId != 0 {
		startWeek, err := weekRepo.FindByID(int(req.StartWeekId))
		if err != nil {
			return nil, err
		}
		semester.StartDate = startWeek.StartDate

		if req.BeginDate == "" {
			parsedBegin, _ := time.Parse("2006-01-02", startWeek.StartDate.Format("2006-01-02"))
			semester.BeginDate = parsedBegin
		}
	}

	if req.EndWeekId != 0 {
		endWeek, err := weekRepo.FindByID(int(req.EndWeekId))
		if err != nil {
			return nil, err
		}
		semester.EndDate = endWeek.EndDate
	}

	s.repo.SetContext(c)

	err := s.repo.Create(semester)
	if err != nil {
		return nil, err
	}

	id := int(semester.ID)
	s.StoreHolidays(c, int64(id), req)
	s.repo.SetPreload([]string{
		"Holidays",
	})
	newSemester, _ := s.repo.FindNewByID(id)

	return newSemester, nil
}

func (s *semesterService) Update(c *gin.Context, req *prot.SemesterRequest) (*models.Semester, error) {
	semesterResource := resources.NewSemesterResource()

	if req.PreviousSemesterId != 0 {
		beginDate := s.GetEarliestStartDateFromPrevious(int(req.PreviousSemesterId), map[int]bool{})
		if beginDate != nil {
			req.BeginDate = beginDate.Format("2006-01-02")
		}
	}

	semester := semesterResource.FormatModelSemester(req)
	weekRepo := repositories.NewWeekRepository()

	if req.StartWeekId != 0 {
		startWeek, err := weekRepo.FindByID(int(req.StartWeekId))
		if err != nil {
			return nil, err
		}
		semester.StartDate = startWeek.StartDate

		if req.BeginDate == "" {
			parsedBegin, _ := time.Parse("2006-01-02", startWeek.StartDate.Format("2006-01-02"))
			semester.BeginDate = parsedBegin
		}
	}

	if req.EndWeekId != 0 {
		endWeek, err := weekRepo.FindByID(int(req.EndWeekId))
		if err != nil {
			return nil, err
		}
		semester.EndDate = endWeek.EndDate
	}

	s.repo.SetContext(c)

	err := s.repo.Update(semester)
	if err != nil {
		return nil, err
	}

	id := int(semester.ID)
	s.StoreHolidays(c, int64(id), req)

	s.repo.SetPreload([]string{
		"Holidays",
	})

	updatedSemester, _ := s.repo.FindNewByID(id)

	return updatedSemester, nil
}

func (s *semesterService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *semesterService) Restore(c *gin.Context, id int) (*models.Semester, error) {
	s.repo.SetContext(c)
	semester, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return semester, nil
}

func (s *semesterService) GetEarliestStartDateFromPrevious(previousID int, visited map[int]bool) *time.Time {
	if visited[previousID] {
		return nil
	}
	visited[previousID] = true

	previous, err := s.repo.FindByID(previousID)
	if err != nil || previous == nil || previous.ID == 0 {
		return nil
	}

	if previous.PreviousSemesterId == 0 {
		return &previous.StartDate
	}

	return s.GetEarliestStartDateFromPrevious(int(previous.PreviousSemesterId), visited)
}

func (s *semesterService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	if byUserParam := c.Query("by_user"); byUserParam != "" {
		if byUserParam == "true" || byUserParam == "1" || byUserParam == "yes" {
			userId := utils.GetCurrentUserId(c)

			ids, _ := s.repo.GetIdsByUserID(userId)

			if len(ids) > 0 {
				strIds := make([]string, len(ids))
				for i, id := range ids {
					strIds[i] = strconv.Itoa(id)
				}
				filter["id"] = "in:" + strings.Join(strIds, ",")
			} else {
				filter["id"] = "in:-1"
			}
		}
	}

	if isCurrent := c.Query("is_current"); isCurrent != "" {
		if isCurrent == "true" || isCurrent == "1" || isCurrent == "yes" {
			filter["start_date"] = "<=:" + time.Now().Format("2006-01-02")
			filter["end_date"] = ">=:" + time.Now().Format("2006-01-02")
		}
	}

	return filter, nil
}

func (s *semesterService) StoreHolidays(c *gin.Context, id int64, req *prot.SemesterRequest) error {
	holidayResource := resources.NewHolidayResource()

	if err := s.repo.DeleteOldHolidays(id, nil); err != nil {
		return err
	}

	parseDate := func(s string) time.Time {
		if s == "" {
			return time.Time{}
		}
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
		return time.Time{}
	}

	for _, value := range req.Holidays {
		holidayModel := holidayResource.ProtToModel(value)
		holidayModel.StartDate = parseDate(value.StartDate)
		holidayModel.EndDate = parseDate(value.EndDate)

		createdHoliday, err := s.repo.UpdateOrCreate(*holidayModel)
		if err != nil {
			return err
		}

		ref := models.SemesterRefHoliday{
			SemesterId: id,
			HolidayId:  createdHoliday.ID,
		}
		if err := s.repo.UpdateOrCreateHoliday(ref); err != nil {
			return err
		}
	}

	return nil
}
