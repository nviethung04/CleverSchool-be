package services

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LessonScheduleService interface {
	GetAll(c *gin.Context) ([]dto.ScheduleGroup, error)
}

type lessonScheduleService struct {
	repo repositories.LessonScheduleRepository
}

func NewLessonScheduleService(repo repositories.LessonScheduleRepository) LessonScheduleService {
	return &lessonScheduleService{repo: repo}
}

func (s *lessonScheduleService) GetAll(c *gin.Context) ([]dto.ScheduleGroup, error) {
	userId := utils.GetCurrentUserId(c)

	filters := make(map[string]interface{})

	if startStr := c.Query("start_date"); startStr != "" {
		if startDate, err := time.Parse("2006-01-02", startStr); err == nil {
			filters["start_date"] = startDate
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		if endDate, err := time.Parse("2006-01-02", endStr); err == nil {
			filters["end_date"] = endDate
		}
	}

	if weekDateStr := c.Query("week_date"); weekDateStr != "" {
		if weekDate, err := time.Parse("2006-01-02", weekDateStr); err == nil {
			filters["week_date"] = weekDate
		}
	}

	if weekIDStr := c.Query("week_id"); weekIDStr != "" {
		if weekID, err := strconv.ParseInt(weekIDStr, 10, 64); err == nil {
			filters["week_id"] = weekID
		}
	}

	if len(filters) == 0 {
		filters["week_date"] = time.Now()
	}

	lessonSchedules, err := s.repo.GetSchedulesByUser(userId, filters)
	if err != nil {
		return nil, err
	}

	return s.GroupSchedulesByWeek(lessonSchedules), nil
}

func (s *lessonScheduleService) GroupSchedulesByWeek(schedules []models.LessonSchedule) []dto.ScheduleGroup {
	groupMap := make(map[string]*dto.ScheduleGroup)

	for _, s := range schedules {
		year := s.Week.Year
		weekId := s.Week.ID
		weekNumber := s.Week.WeekNumber
		startDate := s.Week.StartDate
		endDate := s.Week.EndDate
		date := s.ScheduledDate

		key := fmt.Sprintf("%d-%d", year, weekNumber)
		weekday := date.Weekday().String()

		if _, exists := groupMap[key]; !exists {
			groupMap[key] = &dto.ScheduleGroup{
				WeekId:     int(weekId),
				WeekNumber: weekNumber,
				Year:       year,
				StartDate:  startDate,
				EndDate:    endDate,
				Days:       make(map[string][]models.LessonSchedule),
			}
		}

		groupMap[key].Days[weekday] = append(groupMap[key].Days[weekday], s)
	}

	groups := []dto.ScheduleGroup{}
	for _, g := range groupMap {
		groups = append(groups, *g)
	}

	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Year == groups[j].Year {
			return groups[i].WeekNumber < groups[j].WeekNumber
		}
		return groups[i].Year < groups[j].Year
	})

	return groups
}
