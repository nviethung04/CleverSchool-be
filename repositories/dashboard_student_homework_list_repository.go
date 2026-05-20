package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/config"
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
	// JOIN với homework_ref_lessons theo cả 3 điều kiện: homework_id, lesson_id, và course_id (theo chuẩn API /api/dashboard/report/courses/homeworks)
	homeworkQuery := db.ReplicaDB.Model(&models.Homework{}).
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
		Joins("JOIN lessons l ON hrl.lesson_id = l.id").
		Joins("JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN courses c ON ls.course_id = c.id").
		Where("hrl.assigned_at IS NOT NULL").
		Where("homeworks.deleted_at IS NULL")
	if startDate != nil && endDate != nil {
		homeworkQuery = homeworkQuery.Where("DATE(w.start_date) >= DATE(?) AND DATE(w.end_date) <= DATE(?)", time.Unix(*startDate, 0), time.Unix(*endDate, 0))
	} else if len(weekIDs) > 0 {
		homeworkQuery = homeworkQuery.Where("ls.week_id IN ?", weekIDs)
	}
	if courseID != nil {
		homeworkQuery = homeworkQuery.Where("hrl.course_id = ?", *courseID)
	}

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
		AssignedAt    *int64
		CourseID      int64
		CourseName    string
		LessonID      int64
		LessonTitle   string
	}
	// Lấy assigned_at trực tiếp từ DB (timestamp without timezone, đã lưu thời gian +7)
	// Lấy timestamp raw và convert sang Unix timestamp
	// Vì timestamp trong DB đã là +7, coi nó như UTC+7 khi extract epoch để giữ nguyên giá trị
	if err := homeworkQuery.Select("homeworks.*, w.week_number, w.year, w.start_date as week_start_date, w.end_date as week_end_date, hrl.course_id, c.name as course_name, hrl.lesson_id, l.title as lesson_title, EXTRACT(EPOCH FROM hrl.assigned_at::timestamp AT TIME ZONE 'Asia/Ho_Chi_Minh')::bigint as assigned_at").Scan(&homeworksWithWeek).Error; err != nil {
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
		Rate               string
	}
	var homeworkUsers []HomeworkUserInfo
	db.ReplicaDB.Table("homework_users").
		Select("homework_id, questions_completed, created_at, ratio, rate").
		Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
		Scan(&homeworkUsers)
	homeworkUserMap := make(map[int64]HomeworkUserInfo)
	for _, hu := range homeworkUsers {
		homeworkUserMap[hu.HomeworkID] = hu
	}

	config.Log.Info("homeworkUsers: ", homeworkUsers)

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

		rate := ""
		if ok {
			createdAt = hu.CreatedAt.Unix()
			ratio = hu.Ratio
			rate = hu.Rate
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
		if h.AssignedAt != nil && *h.AssignedAt > 0 {
			assignedAt = *h.AssignedAt
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
			QuestionForm:       h.QuestionForm,
			Rate:               rate,
		})
	}
	return result, total, nil
}
