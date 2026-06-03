package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"time"
)

type DashboardStudentHomeworkListRepository interface {
	GetStudentHomeworkList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error)
}

type dashboardStudentHomeworkListRepository struct{}

func NewDashboardStudentHomeworkListRepository() DashboardStudentHomeworkListRepository {
	return &dashboardStudentHomeworkListRepository{}
}

func (r *dashboardStudentHomeworkListRepository) GetStudentHomeworkList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error) {
	// Lấy danh sách tuần
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
	// Query homeworks bằng GORM query builder, join weeks và courses để lấy thông tin tuần, course, lesson
	homeworkQuery := db.ReplicaDB.Model(&models.Homework{}).
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Joins("JOIN lessons l ON hrl.lesson_id = l.id").
		Joins("JOIN lesson_schedules ls ON l.id = ls.lesson_id AND hrl.course_id = ls.course_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN courses c ON ls.course_id = c.id")
	if startDate != nil && endDate != nil {
		homeworkQuery = homeworkQuery.Where("DATE(w.start_date) >= DATE(?) AND DATE(w.end_date) <= DATE(?)", time.Unix(*startDate, 0), time.Unix(*endDate, 0))
	} else if len(weekIDs) > 0 {
		homeworkQuery = homeworkQuery.Where("ls.week_id IN ?", weekIDs)
	}
	if courseID != nil {
		homeworkQuery = homeworkQuery.Where("ls.course_id = ?", *courseID)
	}

	homeworkQuery = homeworkQuery.Where("homeworks.deleted_at IS NULL")

	// Đếm tổng số
	var total int64
	if err := homeworkQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// Lấy dữ liệu trang
	if limit > 0 {
		homeworkQuery = homeworkQuery.Limit(limit)
	}
	if offset > 0 {
		homeworkQuery = homeworkQuery.Offset(offset)
	}
	homeworkQuery = homeworkQuery.Order("homeworks.created_at DESC")
	// Lấy thêm thông tin tuần, course, lesson
	var homeworksWithWeek []struct {
		models.Homework
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
	if err := homeworkQuery.Select("homeworks.*, w.week_number, w.year, w.start_date as week_start_date, w.end_date as week_end_date, hrl.course_id, c.name as course_name, hrl.lesson_id, l.title as lesson_title, hrl.assigned_at as assigned_at").Scan(&homeworksWithWeek).Error; err != nil {
		return nil, 0, err
	}
	// Lấy homework_users cho user này và các homework này
	var homeworkIDs []int64
	for _, h := range homeworksWithWeek {
		homeworkIDs = append(homeworkIDs, h.ID)
	}
	type HomeworkUserInfo struct {
		HomeworkID         int64
		QuestionsCompleted int64
		CreatedAt          time.Time
		Ratio              float64
	}
	var homeworkUsers []HomeworkUserInfo
	db.ReplicaDB.Table("homework_users").
		Select("homework_id, questions_completed, created_at, ratio").
		Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
		Scan(&homeworkUsers)
	homeworkUserMap := make(map[int64]HomeworkUserInfo)
	for _, hu := range homeworkUsers {
		homeworkUserMap[hu.HomeworkID] = hu
	}

	// Lấy số câu skip cho mỗi homework
	type SkipQuestionInfo struct {
		HomeworkID int64
		Count      int64
	}
	var skipQuestions []SkipQuestionInfo
	db.ReplicaDB.Table("homework_user_skip_questions").
		Select("homework_id, COUNT(*) as count").
		Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
		Group("homework_id").
		Scan(&skipQuestions)
	skipQuestionMap := make(map[int64]int64)
	for _, sq := range skipQuestions {
		skipQuestionMap[sq.HomeworkID] = sq.Count
	}
	// Mapping kết quả
	var result []dto.DashboardStudentHomeworkListItemDTO
	for _, h := range homeworksWithWeek {
		hu, ok := homeworkUserMap[h.ID]
		isSubmitted := ok && hu.CreatedAt.Unix() > 0
		questionsCompleted := int(hu.QuestionsCompleted)
		ratio := 0.0
		createdAt := int64(0)
		if ok {
			createdAt = hu.CreatedAt.Unix()
			ratio = hu.Ratio
		}
		skipCount := int(skipQuestionMap[h.ID])
		weekStart := int64(0)
		weekEnd := int64(0)
		assignedAt := int64(0)
		if !h.WeekStartDate.IsZero() {
			weekStart = h.WeekStartDate.Unix()
		}
		if !h.WeekEndDate.IsZero() {
			weekEnd = h.WeekEndDate.Unix()
		}
		if h.AssignedAt != nil && !h.AssignedAt.IsZero() {
			assignedAt = h.AssignedAt.Unix()
		}
		result = append(result, dto.DashboardStudentHomeworkListItemDTO{
			ID:                 h.ID,
			Name:               h.Name,
			Description:        h.Description,
			CoverImage:         extractCoverImageFromInfo(h.CoverImageInfo),
			CreatedAt:          createdAt,
			IsSubmitted:        isSubmitted,
			TotalQuestions:     int(h.TotalQuestions),
			QuestionsCompleted: questionsCompleted,
			Week:               h.WeekNumber,
			Year:               h.Year,
			WeekStartDate:      weekStart,
			WeekEndDate:        weekEnd,
			CourseID:           h.CourseID,
			CourseName:         h.CourseName,
			LessonID:           h.LessonID,
			LessonTitle:        h.LessonTitle,
			Ratio:              ratio,
			SkipQuestionsCount: skipCount,
			AssignedAt:         assignedAt,
		})
	}
	return result, total, nil
}
