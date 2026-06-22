package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DashboardStudentHomeworkRepository interface {
	GetStudentHomeworkStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentHomeworkRepositoryResult, error)
}

type dashboardStudentHomeworkRepository struct{}

type DashboardStudentHomeworkRepositoryResult struct {
	Overview dto.DashboardStudentHomeworkOverviewDTO
	Chart    []dto.DashboardStudentHomeworkWeekDTO
}

type homeworkWeekStat struct {
	WeekNumber         int
	Year               int
	HomeworkAssigned   int
	HomeworkSubmitted  int
	QuestionsCompleted int
	TotalQuestions     int
	AverageRatio       float64
}

func NewDashboardStudentHomeworkRepository() DashboardStudentHomeworkRepository {
	return &dashboardStudentHomeworkRepository{}
}

func (r *dashboardStudentHomeworkRepository) baseHomeworkQuery(userID int64, courseID *int64) *gorm.DB {
	q := db.ReplicaDB.Table("homework_ref_lessons hrl").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id AND h.deleted_at IS NULL").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID).
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("(hrl.course_id IS NULL OR hrl.course_id = 0 OR hrl.course_id = c.id)")
	if courseID != nil {
		q = q.Where("c.id = ?", *courseID)
	}
	return q
}

func (r *dashboardStudentHomeworkRepository) applyDateFilter(
	query *gorm.DB,
	userID int64,
	startDate, endDate *int64,
) *gorm.DB {
	if startDate == nil || endDate == nil {
		return query
	}
	start := time.Unix(*startDate, 0)
	end := time.Unix(*endDate, 0)
	return query.Where(`(
		(hrl.assigned_at IS NOT NULL AND hrl.assigned_at >= ? AND hrl.assigned_at <= ?)
		OR EXISTS (
			SELECT 1 FROM homework_users hu_filter
			WHERE hu_filter.homework_id = h.id
			  AND hu_filter.user_id = ?
			  AND hu_filter.lesson_id = l.id
			  AND hu_filter.created_at >= ?
			  AND hu_filter.created_at <= ?
		)
	)`, start, end, userID, start, end)
}

func (r *dashboardStudentHomeworkRepository) fetchOverview(
	userID int64,
	courseID *int64,
	startDate, endDate *int64,
) (dto.DashboardStudentHomeworkOverviewDTO, error) {
	type overviewRow struct {
		HomeworkAssigned   int
		HomeworkSubmitted  int
		QuestionsCompleted int
		AverageRatio       float64
	}
	var row overviewRow
	q := r.baseHomeworkQuery(userID, courseID)
	q = r.applyDateFilter(q, userID, startDate, endDate)
	err := q.Select(`
		COUNT(DISTINCT h.id) as homework_assigned,
		COUNT(DISTINCT hu.homework_id) as homework_submitted,
		COALESCE(SUM(hu.questions_completed), 0) as questions_completed,
		COALESCE(AVG(NULLIF(hu.ratio, 0)), 0) as average_ratio
	`).
		Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ? AND hu.lesson_id = l.id", userID).
		Scan(&row).Error
	if err != nil {
		return dto.DashboardStudentHomeworkOverviewDTO{}, err
	}

	var totalQuestions int
	totalQ := r.baseHomeworkQuery(userID, courseID)
	totalQ = r.applyDateFilter(totalQ, userID, startDate, endDate)
	err = totalQ.Select("COALESCE(SUM(DISTINCT h.total_questions), 0)").
		Scan(&totalQuestions).Error
	if err != nil {
		return dto.DashboardStudentHomeworkOverviewDTO{}, err
	}

	percentCompleted := 0.0
	if totalQuestions > 0 {
		percentCompleted = float64(row.QuestionsCompleted) / float64(totalQuestions) * 100
	}
	return dto.DashboardStudentHomeworkOverviewDTO{
		TotalHomeworkAssigned:   row.HomeworkAssigned,
		TotalHomeworkSubmitted:  row.HomeworkSubmitted,
		TotalQuestionsCompleted: row.QuestionsCompleted,
		TotalQuestions:          totalQuestions,
		PercentCompleted:        percentCompleted,
		AverageRatio:            row.AverageRatio,
	}, nil
}

func (r *dashboardStudentHomeworkRepository) GetStudentHomeworkStats(
	userID int64,
	courseID *int64,
	startDate, endDate *int64,
) (*DashboardStudentHomeworkRepositoryResult, error) {
	var courseStart *time.Time
	if courseID != nil {
		var course models.Course
		if err := db.ReplicaDB.Select("start_date").Where("id = ?", *courseID).First(&course).Error; err == nil {
			t := course.StartDate
			courseStart = &t
		}
	}

	overview, err := r.fetchOverview(userID, courseID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var weeks []models.Week
	weekQuery := db.ReplicaDB.Model(&models.Week{})
	if startDate != nil {
		weekQuery = weekQuery.Where("start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		weekQuery = weekQuery.Where("end_date <= ?", time.Unix(*endDate, 0))
	}
	weekQuery = weekQuery.Order("year, week_number")
	if err := weekQuery.Find(&weeks).Error; err != nil {
		return nil, err
	}

	var result []dto.DashboardStudentHomeworkWeekDTO
	if len(weeks) > 0 {
		var stats []homeworkWeekStat
		homeworkQuery := db.ReplicaDB.Table("weeks w").
			Select(`w.week_number, w.year,
				COUNT(DISTINCT CASE WHEN hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND h.deleted_at IS NULL THEN h.id END) as homework_assigned,
				COUNT(DISTINCT hu.homework_id) as homework_submitted,
				COALESCE(SUM(hu.questions_completed),0) as questions_completed,
				COALESCE(SUM(h.total_questions),0) as total_questions,
				COALESCE(AVG(hu.ratio),0) as average_ratio`).
			Joins("LEFT JOIN lesson_schedules ls ON w.id = ls.week_id").
			Joins("LEFT JOIN lessons l ON ls.lesson_id = l.id").
			Joins("LEFT JOIN user_courses uc ON uc.course_id = ls.course_id AND uc.user_id = ?", userID).
			Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND hrl.course_id = ls.course_id").
			Joins("LEFT JOIN homeworks h ON h.id = hrl.homework_id AND h.deleted_at IS NULL").
			Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ? AND hu.lesson_id = l.id", userID)
		if courseID != nil {
			homeworkQuery = homeworkQuery.Where("ls.course_id = ?", *courseID)
		}
		if startDate != nil {
			homeworkQuery = homeworkQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
		}
		if endDate != nil {
			homeworkQuery = homeworkQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
		}
		homeworkQuery = homeworkQuery.Group("w.week_number, w.year, w.start_date, w.end_date").Order("w.year, w.week_number, w.start_date")
		if err := homeworkQuery.Scan(&stats).Error; err != nil {
			return nil, err
		}
		statMap := make(map[string]homeworkWeekStat)
		for _, s := range stats {
			statMap[fmt.Sprintf("%d-%d", s.Year, s.WeekNumber)] = s
		}
		for _, w := range weeks {
			key := fmt.Sprintf("%d-%d", w.Year, w.WeekNumber)
			stat, ok := statMap[key]
			result = append(result, r.buildWeekDTO(w, courseStart, stat, ok))
		}
	} else if overview.TotalHomeworkAssigned > 0 || overview.TotalHomeworkSubmitted > 0 {
		startUnix := int64(0)
		endUnix := int64(0)
		if startDate != nil {
			startUnix = *startDate
		}
		if endDate != nil {
			endUnix = *endDate
		}
		result = append(result, dto.DashboardStudentHomeworkWeekDTO{
			WeekNumber:         1,
			Year:               time.Unix(startUnix, 0).Year(),
			CourseWeek:         1,
			HomeworkSubmitted:  overview.TotalHomeworkSubmitted,
			StartDate:          startUnix,
			EndDate:            endUnix,
			HomeworkAssigned:   overview.TotalHomeworkAssigned,
			QuestionsCompleted: overview.TotalQuestionsCompleted,
			TotalQuestions:     overview.TotalQuestions,
			PercentCompleted:   overview.PercentCompleted,
			AverageRatio:       overview.AverageRatio,
		})
	}

	return &DashboardStudentHomeworkRepositoryResult{
		Overview: overview,
		Chart:    result,
	}, nil
}

func (r *dashboardStudentHomeworkRepository) buildWeekDTO(
	w models.Week,
	courseStart *time.Time,
	stat homeworkWeekStat,
	ok bool,
) dto.DashboardStudentHomeworkWeekDTO {
	percent := 0.0
	courseWeek := 0
	if courseStart != nil {
		weekDate := time.Date(w.StartDate.Year(), w.StartDate.Month(), w.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		courseDate := time.Date(courseStart.Year(), courseStart.Month(), courseStart.Day(), 0, 0, 0, 0, time.UTC)
		days := int(weekDate.Sub(courseDate).Hours() / 24)
		if days >= 0 {
			courseWeek = days/7 + 1
		}
	}
	if !ok {
		return dto.DashboardStudentHomeworkWeekDTO{
			WeekNumber: w.WeekNumber, Year: w.Year, CourseWeek: courseWeek,
			StartDate: w.StartDate.Unix(), EndDate: w.EndDate.Unix(),
		}
	}
	if stat.TotalQuestions > 0 {
		percent = float64(stat.QuestionsCompleted) / float64(stat.TotalQuestions) * 100
	}
	return dto.DashboardStudentHomeworkWeekDTO{
		WeekNumber:         w.WeekNumber,
		Year:               w.Year,
		CourseWeek:         courseWeek,
		HomeworkSubmitted:  stat.HomeworkSubmitted,
		StartDate:          w.StartDate.Unix(),
		EndDate:            w.EndDate.Unix(),
		HomeworkAssigned:   stat.HomeworkAssigned,
		QuestionsCompleted: stat.QuestionsCompleted,
		TotalQuestions:     stat.TotalQuestions,
		PercentCompleted:   percent,
		AverageRatio:       stat.AverageRatio,
	}
}
