package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
)

type DashboardTeacherHomeworkScoredRepository interface {
	GetScoredHomeworks(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherHomeworkScoredListRequest) ([]dto.DashboardTeacherHomeworkScored, int64, error)
}

type dashboardTeacherHomeworkScoredRepository struct{}

func NewDashboardTeacherHomeworkScoredRepository() DashboardTeacherHomeworkScoredRepository {
	return &dashboardTeacherHomeworkScoredRepository{}
}

func (r *dashboardTeacherHomeworkScoredRepository) GetScoredHomeworks(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherHomeworkScoredListRequest) ([]dto.DashboardTeacherHomeworkScored, int64, error) {
	var homeworks []dto.DashboardTeacherHomeworkScored
	var totalCount int64

	query := db.ReplicaDB.Table("homework_users hu").
		Select(`
			DISTINCT ON (hu.lesson_id, hu.homework_id, hu.user_id)
			hu.user_id, hu.homework_id, hu.lesson_id,
			hrl.course_id,
			u.name as student_name,
			h.name as homework_name,
			l.title as lesson_title,
			c.name as course_name,
			hu.created_at as submitted_at,
			hu.ratio,
			h.total_questions,
			0 as unscored_questions,
			CASE 
				WHEN hc.content IS NOT NULL AND hc.content != '' THEN true 
				ELSE false 
			END as has_comment,
			COALESCE(hc.content, '') as comment_content
		`).
		Joins("JOIN users u ON hu.user_id = u.id").
		Joins("LEFT JOIN homeworks h ON hu.homework_id = h.id").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = hu.homework_id AND hrl.lesson_id = hu.lesson_id").
		Joins("LEFT JOIN lessons l ON l.id = hu.lesson_id").
		Joins("LEFT JOIN courses c ON c.id = hrl.course_id").
		Joins("LEFT JOIN homework_comments hc ON hc.homework_id = hu.homework_id AND hc.student_id = hu.user_id AND hc.deleted_at IS NULL").
		Where("hu.status_scoring = ?", 2).
		Where("h.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")

	// Apply filters directly on homework_users table
	if req.HomeworkID > 0 {
		query = query.Where("hu.homework_id = ?", req.HomeworkID)
	}
	if req.UserID > 0 {
		query = query.Where("hu.user_id = ?", req.UserID)
	}
	if req.CourseID > 0 {
		query = query.Where("hrl.course_id = ?", req.CourseID)
	}

	// Nếu chỉ lấy homework cùng khóa (role_id = 2 hoặc 3)
	if onlyUserCourses && userID > 0 {
		query = query.Joins("LEFT JOIN user_courses ON hrl.course_id = user_courses.course_id").
			Where("user_courses.user_id = ?", userID)
	}

	// Filter by start_date and end_date (only date part, not time)
	if req.StartDate > 0 && req.EndDate > 0 {
		// Lọc homework trong khoảng thời gian từ start_date đến end_date
		query = query.Where("DATE(hu.created_at) >= DATE(to_timestamp(?)) AND DATE(hu.created_at) <= DATE(to_timestamp(?))", req.StartDate, req.EndDate)
	} else if req.StartDate > 0 {
		query = query.Where("DATE(hu.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	} else if req.EndDate > 0 {
		query = query.Where("DATE(hu.created_at) <= DATE(to_timestamp(?))", req.EndDate)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		// Với DISTINCT ON, cần ORDER BY theo các field trong DISTINCT ON trước
		query = query.Order("hu.lesson_id, hu.homework_id, hu.user_id")
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		// Với DISTINCT ON, cần ORDER BY theo các field trong DISTINCT ON trước
		query = query.Order("hu.lesson_id, hu.homework_id, hu.user_id, hu.created_at DESC")
	}

	err := query.Find(&homeworks).Error
	return homeworks, totalCount, err
}
