package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"fmt"
	"time"
)

type DashboardStudentExamRepository interface {
	GetStudentExamStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentExamRepositoryResult, error)
}

type dashboardStudentExamRepository struct{}

func NewDashboardStudentExamRepository() DashboardStudentExamRepository {
	return &dashboardStudentExamRepository{}
}

type DashboardStudentExamRepositoryResult struct {
	Overview dto.DashboardStudentExamOverviewDTO
	Chart    []dto.DashboardStudentExamWeekDTO
}

func (r *dashboardStudentExamRepository) GetStudentExamStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentExamRepositoryResult, error) {
	// Lấy danh sách tuần trong khoảng
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
	// Lấy dữ liệu exam theo tuần
	type ExamStat struct {
		WeekNumber    int
		Year          int
		ExamSubmitted int
		AvgScore      float64
		AvgRatio      float64
	}
	var stats []ExamStat
	examQuery := db.ReplicaDB.Table("weeks w").
		Select(`w.week_number, w.year, w.start_date, w.end_date,
			COUNT(eu.id) AS exam_submitted,
			COALESCE(AVG(eu.score), 0) AS avg_score,
			COALESCE(AVG(eu.ratio), 0) AS avg_ratio`).
		Joins("LEFT JOIN lesson_schedules ls ON w.id = ls.week_id").
		Joins("LEFT JOIN lessons l ON ls.lesson_id = l.id").
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.lesson_id = l.id").
		Joins("LEFT JOIN exams e ON e.id = erl.exam_id AND e.deleted_at IS NULL").
		Joins("LEFT JOIN exam_users eu ON eu.exam_id = e.id AND eu.user_id = ?", userID)
	if courseID != nil {
		examQuery = examQuery.Where("ls.course_id = ?", *courseID)
	}
	if startDate != nil {
		examQuery = examQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		examQuery = examQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
	}
	examQuery = examQuery.Group("w.week_number, w.year, w.start_date, w.end_date").Order("w.year, w.week_number, w.start_date")
	if err := examQuery.Scan(&stats).Error; err != nil {
		return nil, err
	}
	// Map dữ liệu exam theo tuần
	statMap := make(map[string]ExamStat)
	for _, s := range stats {
		key := fmt.Sprintf("%d-%d", s.Year, s.WeekNumber)
		statMap[key] = s
	}
	// Lấy dữ liệu exam đã giao theo tuần
	var assignedStats []struct {
		WeekNumber   int
		Year         int
		ExamAssigned int
	}
	assignedQuery := db.ReplicaDB.Table("weeks w").
		Select("w.week_number, w.year, COUNT(DISTINCT e.id) AS exam_assigned").
		Joins("LEFT JOIN lesson_schedules ls ON w.id = ls.week_id").
		Joins("LEFT JOIN lessons l ON ls.lesson_id = l.id").
		Joins("LEFT JOIN chapters ch ON l.chapter_id = ch.id").
		Joins("LEFT JOIN courses c ON ch.program_id = c.program_id").
		Joins("LEFT JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID).
		Joins(`LEFT JOIN exam_ref_lessons erl ON erl.lesson_id = l.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0`).
		Joins("LEFT JOIN exams e ON e.id = erl.exam_id AND e.deleted_at IS NULL")
	if courseID != nil {
		assignedQuery = assignedQuery.Where("ls.course_id = ?", *courseID)
	}
	if startDate != nil {
		assignedQuery = assignedQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		assignedQuery = assignedQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
	}
	assignedQuery = assignedQuery.Group("w.week_number, w.year, w.start_date, w.end_date").Order("w.year, w.week_number, w.start_date")
	if err := assignedQuery.Scan(&assignedStats).Error; err != nil {
		return nil, err
	}
	assignedMap := make(map[string]int)
	for _, a := range assignedStats {
		key := fmt.Sprintf("%d-%d", a.Year, a.WeekNumber)
		assignedMap[key] = a.ExamAssigned
	}
	// Tính tổng exam_assigned
	var totalExamAssigned int64
	assignedTotalQuery := db.ReplicaDB.Table("weeks w").
		Select("COUNT(DISTINCT e.id)").
		Joins("LEFT JOIN lesson_schedules ls ON w.id = ls.week_id").
		Joins("LEFT JOIN lessons l ON ls.lesson_id = l.id").
		Joins("LEFT JOIN chapters ch ON l.chapter_id = ch.id").
		Joins("LEFT JOIN courses c ON ch.program_id = c.program_id").
		Joins("LEFT JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID).
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.lesson_id = l.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Joins("LEFT JOIN exams e ON e.id = erl.exam_id AND e.deleted_at IS NULL")
	if courseID != nil {
		assignedTotalQuery = assignedTotalQuery.Where("ls.course_id = ?", *courseID)
	}
	if startDate != nil {
		assignedTotalQuery = assignedTotalQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		assignedTotalQuery = assignedTotalQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
	}
	if err := assignedTotalQuery.Count(&totalExamAssigned).Error; err != nil {
		return nil, err
	}
	// Tính tổng exam_submitted và avg_ratio toàn bộ
	var totalExamSubmitted int64
	var avgRatio float64
	submittedQuery := db.ReplicaDB.Table("exam_users eu").
		Select("COUNT(eu.id), COALESCE(AVG(eu.ratio),0)").
		Joins("JOIN exams e ON eu.exam_id = e.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons l ON erl.lesson_id = l.id").
		Joins("JOIN lesson_schedules ls ON l.id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN chapters ch ON l.chapter_id = ch.id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID)
	if courseID != nil {
		submittedQuery = submittedQuery.Where("ls.course_id = ?", *courseID)
	}
	if startDate != nil {
		submittedQuery = submittedQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		submittedQuery = submittedQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
	}
	if err := submittedQuery.Row().Scan(&totalExamSubmitted, &avgRatio); err != nil {
		return nil, err
	}
	// Ghép dữ liệu tuần với dữ liệu exam
	var result []dto.DashboardStudentExamWeekDTO
	for _, w := range weeks {
		key := fmt.Sprintf("%d-%d", w.Year, w.WeekNumber)
		stat, ok := statMap[key]
		examAssigned := assignedMap[key]
		if ok {
			result = append(result, dto.DashboardStudentExamWeekDTO{
				WeekNumber:    w.WeekNumber,
				Year:          w.Year,
				ExamSubmitted: stat.ExamSubmitted,
				AvgScore:      stat.AvgScore,
				StartDate:     w.StartDate.Unix(),
				EndDate:       w.EndDate.Unix(),
				AvgRatio:      stat.AvgRatio,
				ExamAssigned:  examAssigned,
			})
		} else {
			result = append(result, dto.DashboardStudentExamWeekDTO{
				WeekNumber:    w.WeekNumber,
				Year:          w.Year,
				ExamSubmitted: 0,
				AvgScore:      0,
				StartDate:     w.StartDate.Unix(),
				EndDate:       w.EndDate.Unix(),
				AvgRatio:      0,
				ExamAssigned:  examAssigned,
			})
		}
	}
	overview := dto.DashboardStudentExamOverviewDTO{
		TotalExamAssigned:  int(totalExamAssigned),
		TotalExamSubmitted: int(totalExamSubmitted),
		AvgRatio:           avgRatio,
	}
	return &DashboardStudentExamRepositoryResult{
		Overview: overview,
		Chart:    result,
	}, nil
}
