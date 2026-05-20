package services

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LessonScheduleService interface {
	GetAll(c *gin.Context) ([]dto.ScheduleGroup, error)
	SyncAllProgram(courseID int64) (*prot.LessonScheduleSyncResponse, error)
}

type lessonScheduleService struct {
	repo        repositories.LessonScheduleRepository
	copyService LessonScheduleCopySharedService
}

func NewLessonScheduleService(repo repositories.LessonScheduleRepository) LessonScheduleService {
	return &lessonScheduleService{
		repo:        repo,
		copyService: NewLessonScheduleCopySharedService(),
	}
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

func (s *lessonScheduleService) SyncAllProgram(courseID int64) (*prot.LessonScheduleSyncResponse, error) {
	if courseID <= 0 {
		return nil, errors.New("course_id phải lớn hơn 0")
	}

	// Lấy program_id từ course
	var courseInfo struct {
		ID        int64
		ProgramID int64
	}
	err := db.ReplicaDB.Table("courses").
		Select("id, program_id").
		Where("id = ? AND deleted_at IS NULL", courseID).
		Scan(&courseInfo).Error
	if err != nil {
		return nil, fmt.Errorf("không thể lấy thông tin khóa học: %w", err)
	}

	if courseInfo.ID == 0 {
		return nil, errors.New("không tìm thấy khóa học")
	}

	if courseInfo.ProgramID == 0 {
		return nil, errors.New("khóa học không có program_id")
	}

	// Tìm tất cả courses cùng program_id (trừ course nguồn)
	var targetCourseIDs []int64
	err = db.ReplicaDB.Table("courses").
		Select("id").
		Where("program_id = ? AND id != ? AND deleted_at IS NULL", courseInfo.ProgramID, courseID).
		Pluck("id", &targetCourseIDs).Error
	if err != nil {
		return nil, fmt.Errorf("không thể lấy danh sách khóa học: %w", err)
	}

	if len(targetCourseIDs) == 0 {
		return nil, errors.New("không có khóa học nào khác cùng program để đồng bộ")
	}

	// Copy lịch học từ course nguồn sang tất cả courses khác
	var syncedCourseIDs []int64
	var failedCourses []int64
	for _, targetID := range targetCourseIDs {
		copyResult, err := s.copyService.CopyLessonSchedulesBetweenCourses(courseID, targetID)
		if err != nil {
			failedCourses = append(failedCourses, targetID)
			continue
		}
		if !copyResult.Success {
			failedCourses = append(failedCourses, targetID)
			continue
		}
		syncedCourseIDs = append(syncedCourseIDs, targetID)
	}

	if len(syncedCourseIDs) == 0 {
		return nil, fmt.Errorf("không thể đồng bộ sang bất kỳ khóa học nào. Có %d khóa thất bại", len(failedCourses))
	}

	message := fmt.Sprintf("Đồng bộ thành công sang %d khóa học", len(syncedCourseIDs))
	if len(failedCourses) > 0 {
		message += fmt.Sprintf(", %d khóa thất bại", len(failedCourses))
	}

	return &prot.LessonScheduleSyncResponse{
		Success:         true,
		Message:         message,
		SourceCourseId:  courseID,
		SyncedCourseIds: syncedCourseIDs,
		FailedCourseIds: failedCourses,
		SyncedCount:     int32(len(syncedCourseIDs)),
		FailedCount:     int32(len(failedCourses)),
	}, nil
}
