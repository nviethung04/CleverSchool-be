package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"time"
)

type DashboardStudentExamListRepository interface {
	GetStudentExamList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentExamListItemDTO, int64, error)
}

type dashboardStudentExamListRepository struct{}

func NewDashboardStudentExamListRepository() DashboardStudentExamListRepository {
	return &dashboardStudentExamListRepository{}
}

func (r *dashboardStudentExamListRepository) GetStudentExamList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentExamListItemDTO, int64, error) {
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

	// Query exams bằng GORM query builder, join weeks và courses để lấy thông tin tuần, course, lesson
	examQuery := db.ReplicaDB.Model(&models.Exam{}).
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Joins("JOIN lessons l ON erl.lesson_id = l.id").
		Joins("JOIN lesson_schedules ls ON l.id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN courses c ON ls.course_id = c.id")
	if startDate != nil && endDate != nil {
		examQuery = examQuery.Where("DATE(w.start_date) >= DATE(?) AND DATE(w.end_date) <= DATE(?)", time.Unix(*startDate, 0), time.Unix(*endDate, 0))
	} else if len(weekIDs) > 0 {
		examQuery = examQuery.Where("ls.week_id IN ?", weekIDs)
	}
	if courseID != nil {
		examQuery = examQuery.Where("ls.course_id = ?", *courseID)
	}
	examQuery = examQuery.Where("exams.deleted_at IS NULL")
	// Đếm tổng số
	var total int64
	if err := examQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// Lấy dữ liệu trang
	if limit > 0 {
		examQuery = examQuery.Limit(limit)
	}
	if offset > 0 {
		examQuery = examQuery.Offset(offset)
	}
	examQuery = examQuery.Order("exams.deadline DESC")
	// Lấy thêm thông tin tuần, course, lesson
	var examsWithWeek []struct {
		models.Exam
		WeekNumber    int
		Year          int
		WeekStartDate time.Time
		WeekEndDate   time.Time
		CourseName    string
		LessonTitle   string
	}
	if err := examQuery.Select("exams.*, w.week_number, w.year, w.start_date as week_start_date, w.end_date as week_end_date, c.name as course_name, l.title as lesson_title").Scan(&examsWithWeek).Error; err != nil {
		return nil, 0, err
	}
	// Lấy danh sách exam_id
	var examIDs []int64
	for _, e := range examsWithWeek {
		examIDs = append(examIDs, e.ID)
	}
	// Lấy toàn bộ exam_users cho user này và các exam này
	type ExamUserInfo struct {
		ExamID    int64
		Ratio     float64
		Duration  int64
		CreatedAt time.Time
	}
	var examUsers []ExamUserInfo
	db.ReplicaDB.Table("exam_users").
		Select("exam_id, ratio, time as duration, created_at").
		Where("exam_id IN ? AND user_id = ?", examIDs, userID).
		Scan(&examUsers)
	examUserMap := make(map[int64]ExamUserInfo)
	for _, eu := range examUsers {
		examUserMap[eu.ExamID] = eu
	}
	// Lấy toàn bộ unscored_count cho user này và các exam này
	type Unscored struct {
		ExamID int64
		Count  int64
	}
	var unscoredList []Unscored
	db.ReplicaDB.Table("exam_question_user_manual_scoring").
		Select("exam_id, COUNT(*) as count").
		Where("exam_id IN ? AND user_id = ? AND is_scored = false", examIDs, userID).
		Group("exam_id").
		Scan(&unscoredList)
	unscoredMap := make(map[int64]int64)
	for _, u := range unscoredList {
		unscoredMap[u.ExamID] = u.Count
	}
	// Mapping kết quả
	var result []dto.DashboardStudentExamListItemDTO
	for _, e := range examsWithWeek {
		eu, ok := examUserMap[e.ID]
		isSubmitted := ok && eu.CreatedAt.Unix() > 0
		ratio := float64(0)
		duration := int64(0)
		createdAt := int64(0)
		if ok {
			ratio = eu.Ratio
			duration = eu.Duration
			createdAt = eu.CreatedAt.Unix()
		}
		unscored := unscoredMap[e.ID]
		result = append(result, dto.DashboardStudentExamListItemDTO{
			ID:            e.ID,
			Name:          e.Name,
			Description:   e.Description,
			CoverImage:    extractCoverImageFromInfo(e.CoverImageInfo),
			Deadline:      e.Deadline.Unix(),
			IsSubmitted:   isSubmitted,
			UnscoredCount: int(unscored),
			Ratio:         ratio,
			Duration:      duration,
			CreatedAt:     createdAt,
			Week:          e.WeekNumber,
			Year:          e.Year,
			WeekStartDate: e.WeekStartDate.Unix(),
			WeekEndDate:   e.WeekEndDate.Unix(),
			CourseName:    e.CourseName,
			LessonTitle:   e.LessonTitle,
		})
	}
	return result, total, nil
}

// extractCoverImageFromInfo: parse cover_image_info jsonb nếu cần
func extractCoverImageFromInfo(info interface{}) string {
	// TODO: parse JSON nếu cần, hoặc trả về rỗng
	return ""
}
