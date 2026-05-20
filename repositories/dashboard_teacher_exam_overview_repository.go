package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/requests"

	"gorm.io/gorm"
)

type DashboardTeacherExamOverviewRepository interface {
	GetOverview(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherExamOverviewRequest) (*dto.DashboardTeacherExamOverview, error)
}

type dashboardTeacherExamOverviewRepository struct{}

func NewDashboardTeacherExamOverviewRepository() DashboardTeacherExamOverviewRepository {
	return &dashboardTeacherExamOverviewRepository{}
}

func (r *dashboardTeacherExamOverviewRepository) GetOverview(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherExamOverviewRequest) (*dto.DashboardTeacherExamOverview, error) {
	var overview dto.DashboardTeacherExamOverview

	// Helper function to create base query with all joins and filters
	createBaseQuery := func() *gorm.DB {
		query := db.ReplicaDB.Table("exam_users").
			Joins("LEFT JOIN exams ON exam_users.exam_id = exams.id").
			Joins("LEFT JOIN exam_ref_lessons ON exam_ref_lessons.exam_id = exams.id").
			Joins("LEFT JOIN lessons ON exam_ref_lessons.lesson_id = lessons.id").
			Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
			Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
			Joins("LEFT JOIN subjects ON courses.subject_id = subjects.id")

		// Apply filters
		if req.CourseID > 0 {
			query = query.Where("courses.id = ?", req.CourseID)
		}
		if req.ExamID > 0 {
			query = query.Where("exams.id = ?", req.ExamID)
		}
		if req.StartDate > 0 {
			query = query.Where("DATE(exam_users.created_at) >= DATE(to_timestamp(?))", req.StartDate)
		}
		if req.EndDate > 0 {
			query = query.Where("DATE(exam_users.created_at) <= DATE(to_timestamp(?))", req.EndDate)
		}

		// Nếu chỉ lấy exam cùng khóa (role_id = 2 hoặc 3)
		if onlyUserCourses && userID > 0 {
			query = query.Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
				Where("user_courses.user_id = ?", userID)
		}

		return query
	}

	// Count total submitted (distinct exam_users to avoid duplicates from multiple lessons)
	var totalSubmitted int64
	baseQuery := createBaseQuery()
	if err := baseQuery.Select("COUNT(DISTINCT exam_users.id)").Scan(&totalSubmitted).Error; err != nil {
		return nil, err
	}
	overview.TotalSubmitted = totalSubmitted

	// Count total scored (no unscored questions)
	scoredQuery := createBaseQuery()
	if err := scoredQuery.Where("NOT EXISTS (SELECT 1 FROM exam_question_user_manual_scoring equms WHERE equms.exam_id = exam_users.exam_id AND equms.user_id = exam_users.user_id AND equms.is_scored = false)").
		Select("COUNT(DISTINCT exam_users.id)").Scan(&overview.TotalScored).Error; err != nil {
		return nil, err
	}

	// Count total unscored (has unscored questions)
	unscoredQuery := createBaseQuery()
	if err := unscoredQuery.Where("EXISTS (SELECT 1 FROM exam_question_user_manual_scoring equms WHERE equms.exam_id = exam_users.exam_id AND equms.user_id = exam_users.user_id AND equms.is_scored = false)").
		Select("COUNT(DISTINCT exam_users.id)").Scan(&overview.TotalUnscored).Error; err != nil {
		return nil, err
	}

	// Calculate completion rate
	if overview.TotalSubmitted > 0 {
		overview.CompletionRate = float64(overview.TotalScored) / float64(overview.TotalSubmitted) * 100
	} else {
		overview.CompletionRate = 0
	}

	return &overview, nil
}
