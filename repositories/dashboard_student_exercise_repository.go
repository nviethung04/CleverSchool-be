package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"fmt"
	"time"
)

type DashboardStudentExerciseRepository interface {
	GetStudentExerciseStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentHomeworkRepositoryResult, error)
}

type dashboardStudentExerciseRepository struct{}

func NewDashboardStudentExerciseRepository() DashboardStudentExerciseRepository {
	return &dashboardStudentExerciseRepository{}
}

func (r *dashboardStudentExerciseRepository) GetStudentExerciseStats(userID int64, courseID *int64, startDate, endDate *int64) (*DashboardStudentHomeworkRepositoryResult, error) {
	var courseStart *time.Time
	if courseID != nil {
		var course models.Course
		if err := db.ReplicaDB.Select("start_date").Where("id = ?", *courseID).First(&course).Error; err == nil {
			t := course.StartDate
			courseStart = &t
		}
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

	type ExerciseStat struct {
		WeekNumber         int
		Year               int
		ExerciseAssigned   int
		ExerciseSubmitted  int
		QuestionsCompleted int
		TotalQuestions     int
		AverageRatio       float64
	}
	var stats []ExerciseStat
	exerciseQuery := db.ReplicaDB.Table("weeks w").
		Select(`w.week_number, w.year,
            COUNT(DISTINCT CASE WHEN erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND e.deleted_at IS NULL THEN e.id END) as exercise_assigned,
            COUNT(DISTINCT eu.exercise_id) as exercise_submitted,
            COALESCE(SUM(qc.questions_completed), 0) as questions_completed,
            COALESCE(SUM(e.total_questions), 0) as total_questions,
            COALESCE(AVG(eu.ratio), 0) as average_ratio`).
		Joins("LEFT JOIN lesson_schedules ls ON w.id = ls.week_id").
		Joins("LEFT JOIN lessons l ON ls.lesson_id = l.id").
		Joins("LEFT JOIN user_courses uc ON uc.course_id = ls.course_id AND uc.user_id = ?", userID).
		Joins("LEFT JOIN exercise_ref_lessons erl ON erl.lesson_id = l.id AND erl.assigned_at IS NOT NULL AND erl.course_id = ls.course_id").
		Joins("LEFT JOIN exercises e ON e.id = erl.exercise_id AND e.deleted_at IS NULL").
		Joins("LEFT JOIN exercise_users eu ON eu.exercise_id = e.id AND eu.user_id = ?", userID).
		Joins(`LEFT JOIN (
            SELECT exercise_id, user_id, COUNT(DISTINCT question_id) AS questions_completed
            FROM exercise_question_users
            GROUP BY exercise_id, user_id
        ) qc ON qc.exercise_id = e.id AND qc.user_id = ?`, userID)

	if courseID != nil {
		exerciseQuery = exerciseQuery.Where("ls.course_id = ?", *courseID)
	}
	if startDate != nil {
		exerciseQuery = exerciseQuery.Where("w.start_date >= ?", time.Unix(*startDate, 0))
	}
	if endDate != nil {
		exerciseQuery = exerciseQuery.Where("w.end_date <= ?", time.Unix(*endDate, 0))
	}
	exerciseQuery = exerciseQuery.Group("w.week_number, w.year, w.start_date, w.end_date").Order("w.year, w.week_number, w.start_date")
	if err := exerciseQuery.Scan(&stats).Error; err != nil {
		return nil, err
	}

	statMap := make(map[string]ExerciseStat)
	for _, s := range stats {
		key := fmt.Sprintf("%d-%d", s.Year, s.WeekNumber)
		statMap[key] = s
	}

	var totalAssigned, totalSubmitted, totalQuestionsCompleted, totalQuestions int
	var totalAverageRatio float64
	var weekCount int
	for _, s := range stats {
		totalAssigned += s.ExerciseAssigned
		totalSubmitted += s.ExerciseSubmitted
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
		TotalHomeworkAssigned:   totalAssigned,
		TotalHomeworkSubmitted:  totalSubmitted,
		TotalQuestionsCompleted: totalQuestionsCompleted,
		TotalQuestions:          totalQuestions,
		PercentCompleted:        percentCompleted,
		AverageRatio:            averageRatio,
	}

	var result []dto.DashboardStudentHomeworkWeekDTO
	for _, w := range weeks {
		key := fmt.Sprintf("%d-%d", w.Year, w.WeekNumber)
		stat, ok := statMap[key]
		percent := 0.0
		if stat.TotalQuestions > 0 {
			percent = float64(stat.QuestionsCompleted) / float64(stat.TotalQuestions) * 100
		}
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
				HomeworkSubmitted:  stat.ExerciseSubmitted,
				StartDate:          w.StartDate.Unix(),
				EndDate:            w.EndDate.Unix(),
				HomeworkAssigned:   stat.ExerciseAssigned,
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
