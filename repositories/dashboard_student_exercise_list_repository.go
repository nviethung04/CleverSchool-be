package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"time"
)

type DashboardStudentExerciseListRepository interface {
	GetStudentExerciseList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error)
}

type dashboardStudentExerciseListRepository struct{}

func NewDashboardStudentExerciseListRepository() DashboardStudentExerciseListRepository {
	return &dashboardStudentExerciseListRepository{}
}

func (r *dashboardStudentExerciseListRepository) GetStudentExerciseList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error) {
	var weekIDs []int64
	if startDate != nil && endDate != nil {
		var weeks []models.Week
		weekQuery := db.ReplicaDB.Model(&models.Week{}).
			Where("start_date >= ?", time.Unix(*startDate, 0)).
			Where("end_date <= ?", time.Unix(*endDate, 0))
		if err := weekQuery.Find(&weeks).Error; err != nil {
			return nil, 0, err
		}
		for _, w := range weeks {
			weekIDs = append(weekIDs, w.ID)
		}
	}

	exerciseQuery := db.ReplicaDB.Model(&models.Exercise{}).
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Joins("JOIN lessons l ON erl.lesson_id = l.id").
		Joins("JOIN lesson_schedules ls ON l.id = ls.lesson_id AND erl.course_id = ls.course_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN courses c ON ls.course_id = c.id")
	if startDate != nil && endDate != nil {
		exerciseQuery = exerciseQuery.Where("DATE(w.start_date) >= DATE(?) AND DATE(w.end_date) <= DATE(?)", time.Unix(*startDate, 0), time.Unix(*endDate, 0))
	} else if len(weekIDs) > 0 {
		exerciseQuery = exerciseQuery.Where("ls.week_id IN ?", weekIDs)
	}
	if courseID != nil {
		exerciseQuery = exerciseQuery.Where("ls.course_id = ?", *courseID)
	}
	exerciseQuery = exerciseQuery.Where("exercises.deleted_at IS NULL")

	var total int64
	if err := exerciseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit > 0 {
		exerciseQuery = exerciseQuery.Limit(limit)
	}
	if offset > 0 {
		exerciseQuery = exerciseQuery.Offset(offset)
	}
	exerciseQuery = exerciseQuery.Order("exercises.created_at DESC")

	var exercisesWithWeek []struct {
		models.Exercise
		WeekNumber    int
		Year          int
		WeekStartDate time.Time
		WeekEndDate   time.Time
		AssignedAt    *time.Time
		CourseID      int64
		CourseName    string
		LessonID      int64
		LessonTitle   string
	}
	if err := exerciseQuery.Select("exercises.*, w.week_number, w.year, w.start_date as week_start_date, w.end_date as week_end_date, erl.course_id, c.name as course_name, erl.lesson_id, l.title as lesson_title, erl.assigned_at as assigned_at").Scan(&exercisesWithWeek).Error; err != nil {
		return nil, 0, err
	}

	var exerciseIDs []int64
	for _, e := range exercisesWithWeek {
		exerciseIDs = append(exerciseIDs, e.ID)
	}

	type ExerciseUserInfo struct {
		ExerciseID int64
		CreatedAt  time.Time
		Ratio      float64
	}
	var exerciseUsers []ExerciseUserInfo
	if len(exerciseIDs) > 0 {
		db.ReplicaDB.Table("exercise_users").
			Select("exercise_id, created_at, COALESCE(ratio, 0) as ratio").
			Where("exercise_id IN ? AND user_id = ?", exerciseIDs, userID).
			Scan(&exerciseUsers)
	}
	exerciseUserMap := make(map[int64]ExerciseUserInfo)
	for _, eu := range exerciseUsers {
		exerciseUserMap[eu.ExerciseID] = eu
	}

	type QuestionCountInfo struct {
		ExerciseID int64
		Count      int64
	}
	var questionCounts []QuestionCountInfo
	if len(exerciseIDs) > 0 {
		db.ReplicaDB.Table("exercise_question_users").
			Select("exercise_id, COUNT(DISTINCT question_id) as count").
			Where("exercise_id IN ? AND user_id = ?", exerciseIDs, userID).
			Group("exercise_id").
			Scan(&questionCounts)
	}
	questionCountMap := make(map[int64]int64)
	for _, qc := range questionCounts {
		questionCountMap[qc.ExerciseID] = qc.Count
	}

	var result []dto.DashboardStudentHomeworkListItemDTO
	for _, e := range exercisesWithWeek {
		eu, ok := exerciseUserMap[e.ID]
		isSubmitted := ok && eu.CreatedAt.Unix() > 0
		questionsCompleted := int(questionCountMap[e.ID])
		ratio := 0.0
		createdAt := int64(0)
		if ok {
			createdAt = eu.CreatedAt.Unix()
			ratio = eu.Ratio
		}
		weekStart := int64(0)
		weekEnd := int64(0)
		assignedAt := int64(0)
		if !e.WeekStartDate.IsZero() {
			weekStart = e.WeekStartDate.Unix()
		}
		if !e.WeekEndDate.IsZero() {
			weekEnd = e.WeekEndDate.Unix()
		}
		if e.AssignedAt != nil && !e.AssignedAt.IsZero() {
			assignedAt = e.AssignedAt.Unix()
		}
		result = append(result, dto.DashboardStudentHomeworkListItemDTO{
			ID:                 e.ID,
			Name:               e.Name,
			Description:        e.Description,
			CoverImage:         extractCoverImageFromInfo(e.CoverImageInfo),
			CreatedAt:          createdAt,
			IsSubmitted:        isSubmitted,
			TotalQuestions:     int(e.TotalQuestions),
			QuestionsCompleted: questionsCompleted,
			Week:               e.WeekNumber,
			Year:               e.Year,
			WeekStartDate:      weekStart,
			WeekEndDate:        weekEnd,
			CourseID:           e.CourseID,
			CourseName:         e.CourseName,
			LessonID:           e.LessonID,
			LessonTitle:        e.LessonTitle,
			Ratio:              ratio,
			SkipQuestionsCount: 0,
			AssignedAt:         assignedAt,
		})
	}
	return result, total, nil
}
