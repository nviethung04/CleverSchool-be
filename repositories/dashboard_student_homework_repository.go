package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"fmt"
	"time"
)

type DashboardStudentHomeworkRepository interface {
	GetStudentHomeworkStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentHomeworkRepositoryResult, error)
}

type dashboardStudentHomeworkRepository struct{}

type DashboardStudentHomeworkRepositoryResult struct {
	Overview dto.DashboardStudentHomeworkOverviewDTO
	Chart    []dto.DashboardStudentHomeworkWeekDTO
}

func NewDashboardStudentHomeworkRepository() DashboardStudentHomeworkRepository {
	return &dashboardStudentHomeworkRepository{}
}

func (r *dashboardStudentHomeworkRepository) GetStudentHomeworkStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentHomeworkRepositoryResult, error) {
	// Lấy start_date của course (nếu có course_id)
	var courseStart *time.Time
	if courseID != nil {
		var course models.Course
		if err := db.ReplicaDB.Select("start_date").Where("id = ?", *courseID).First(&course).Error; err == nil {
			// create a copy to take address
			t := course.StartDate
			courseStart = &t
		}
	}
	// Lấy danh sách tuần
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
	// Lấy dữ liệu homework đã giao theo tuần
	type HomeworkStat struct {
		WeekNumber         int
		Year               int
		HomeworkAssigned   int
		HomeworkSubmitted  int
		QuestionsCompleted int
		TotalQuestions     int
		AverageRatio       float64
	}
	var stats []HomeworkStat
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
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id AND hrl.assigned_at IS NOT NULL AND hrl.course_id = ls.course_id").
		Joins("LEFT JOIN homeworks h ON h.id = hrl.homework_id AND h.deleted_at IS NULL").
		Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ?", userID)

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
	statMap := make(map[string]HomeworkStat)
	for _, s := range stats {
		key := fmt.Sprintf("%d-%d", s.Year, s.WeekNumber)
		statMap[key] = s
	}
	// Tổng hợp overview
	var totalHomeworkAssigned, totalHomeworkSubmitted, totalQuestionsCompleted, totalQuestions int
	var totalAverageRatio float64
	var weekCount int
	for _, s := range stats {
		totalHomeworkAssigned += s.HomeworkAssigned
		totalHomeworkSubmitted += s.HomeworkSubmitted
		totalQuestionsCompleted += s.QuestionsCompleted
		totalQuestions += s.TotalQuestions
		if s.AverageRatio > 0 {
			totalAverageRatio += s.AverageRatio
			weekCount++
		}
	}
	percentCompleted := 0.0
	if totalQuestions > 0 {
		percentCompleted = float64(totalQuestionsCompleted) / float64(totalQuestions) * 100
	}
	averageRatio := 0.0
	if weekCount > 0 {
		averageRatio = totalAverageRatio / float64(weekCount)
	}
	overview := dto.DashboardStudentHomeworkOverviewDTO{
		TotalHomeworkAssigned:   totalHomeworkAssigned,
		TotalHomeworkSubmitted:  totalHomeworkSubmitted,
		TotalQuestionsCompleted: totalQuestionsCompleted,
		TotalQuestions:          totalQuestions,
		PercentCompleted:        percentCompleted,
		AverageRatio:            averageRatio,
	}
	// Ghép dữ liệu tuần
	var result []dto.DashboardStudentHomeworkWeekDTO
	for _, w := range weeks {
		key := fmt.Sprintf("%d-%d", w.Year, w.WeekNumber)
		stat, ok := statMap[key]
		percent := 0.0
		if stat.TotalQuestions > 0 {
			percent = float64(stat.QuestionsCompleted) / float64(stat.TotalQuestions) * 100
		}
		// Tính course_week = floor((week.start_date - course.start_date)/7) + 1, chuẩn hóa theo ngày UTC
		courseWeek := 0
		if courseStart != nil {
			weekDate := time.Date(w.StartDate.Year(), w.StartDate.Month(), w.StartDate.Day(), 0, 0, 0, 0, time.UTC)
			courseDate := time.Date(courseStart.Year(), courseStart.Month(), courseStart.Day(), 0, 0, 0, 0, time.UTC)
			days := int(weekDate.Sub(courseDate).Hours() / 24)
			if days >= 0 {
				courseWeek = days/7 + 1
			}
		}
		if ok {
			result = append(result, dto.DashboardStudentHomeworkWeekDTO{
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
			})
		} else {
			result = append(result, dto.DashboardStudentHomeworkWeekDTO{
				WeekNumber:         w.WeekNumber,
				Year:               w.Year,
				CourseWeek:         courseWeek,
				HomeworkSubmitted:  0,
				StartDate:          w.StartDate.Unix(),
				EndDate:            w.EndDate.Unix(),
				HomeworkAssigned:   0,
				QuestionsCompleted: 0,
				TotalQuestions:     0,
				PercentCompleted:   0,
				AverageRatio:       0,
			})
		}
	}
	return &DashboardStudentHomeworkRepositoryResult{
		Overview: overview,
		Chart:    result,
	}, nil
}

