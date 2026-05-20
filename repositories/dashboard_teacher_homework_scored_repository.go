package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/requests"
	"encoding/json"
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

	// Nếu không truyền course_id, lấy danh sách course_id của user từ bảng user_courses
	var userCourseIDs []int64
	if req.CourseID == 0 && userID > 0 {
		if err := db.ReplicaDB.Table("user_courses").
			Select("course_id").
			Where("user_id = ?", userID).
			Scan(&userCourseIDs).Error; err != nil {
			return nil, 0, err
		}

		// Nếu user không có khóa nào thì trả về rỗng luôn
		if len(userCourseIDs) == 0 {
			return []dto.DashboardTeacherHomeworkScored{}, 0, nil
		}
	}

	// Bước 1: Lấy danh sách homework_users cơ bản
	query := db.ReplicaDB.Table("homework_users hu").
		Select(`
			DISTINCT ON (hu.lesson_id, hu.homework_id, hu.user_id)
			hu.user_id,
			hu.homework_id,
			hu.lesson_id,
			hrl.course_id,
			u.name as student_name,
			h.name as homework_name,
			l.title as lesson_title,
			c.name as course_name,
			c.object_title as object_title,
			hu.updated_at as submitted_at,
			hu.ratio,
			h.total_questions,
			0 as unscored_questions,
			CASE 
				WHEN hc.content IS NOT NULL AND hc.content != '' THEN true 
				ELSE false 
			END as has_comment,
			COALESCE(hc.content, '') as comment_content,
			COALESCE(graders.graders_json, '[]') as graders_json
		`).
		Joins("JOIN users u ON hu.user_id = u.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("JOIN homeworks h ON hu.homework_id = h.id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = hu.homework_id AND hrl.lesson_id = hu.lesson_id").
		Joins("JOIN lessons l ON l.id = hu.lesson_id").
		Joins("JOIN courses c ON c.id = hrl.course_id").
		Joins("JOIN lesson_schedules ls ON ls.course_id = hrl.course_id AND ls.lesson_id = hrl.lesson_id").
		Joins("JOIN weeks w ON w.id = ls.week_id").
		Joins("JOIN user_courses student_courses ON student_courses.user_id = hu.user_id AND student_courses.course_id = hrl.course_id").
		Joins(`
			LEFT JOIN LATERAL (
				SELECT json_agg(DISTINCT jsonb_build_object(
					'user_id', graders.id,
					'username', graders.username,
					'name', graders.name
				)) AS graders_json
				FROM homework_question_user_manual_scoring hqms
				LEFT JOIN users AS graders ON graders.id = hqms.scoring_by
				WHERE hqms.homework_id = hu.homework_id
				  AND hqms.lesson_id = hu.lesson_id
				  AND hqms.user_id = hu.user_id
			) AS graders ON true`).
		Joins("LEFT JOIN homework_comments hc ON hc.homework_id = hu.homework_id AND hc.student_id = hu.user_id AND hc.deleted_at IS NULL").
		Where("hu.status_scoring = ?", 2).
		Where("h.deleted_at IS NULL").
		Where("u.deleted_at IS NULL").
		Where("urr.role_id = ?", 3)

	// Filter theo course_id
	if req.CourseID > 0 {
		query = query.
			Where("student_courses.course_id = ?", req.CourseID).
			Where("hrl.course_id = ?", req.CourseID)
	} else if len(userCourseIDs) > 0 {
		query = query.
			Where("student_courses.course_id IN (?)", userCourseIDs).
			Where("hrl.course_id IN (?)", userCourseIDs)
	}

	// Apply filters directly on homework_users table
	if req.HomeworkID > 0 {
		query = query.Where("hu.homework_id = ?", req.HomeworkID)
	}
	if req.UserID > 0 {
		query = query.Where("hu.user_id = ?", req.UserID)
	}
	if req.StudentName != "" {
		query = query.Where("unaccent(u.name) ILIKE unaccent(?)", "%"+req.StudentName+"%")
	}
	if req.StartDate > 0 && req.EndDate > 0 {
		query = query.Where("hu.updated_at BETWEEN to_timestamp(?) AND to_timestamp(?)", req.StartDate, req.EndDate)
	} else if req.StartDate > 0 {
		query = query.Where("hu.updated_at >= to_timestamp(?)", req.StartDate)
	} else if req.EndDate > 0 {
		query = query.Where("hu.updated_at <= to_timestamp(?)", req.EndDate)
	}
	// Nếu chỉ lấy homework cùng khóa (role_id = 2 hoặc 3)
	if onlyUserCourses && userID > 0 {
		query = query.Joins("JOIN user_courses teacher_courses ON teacher_courses.course_id = hrl.course_id").
			Where("teacher_courses.user_id = ?", userID)
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
	if err != nil {
		return nil, 0, err
	}

	for i := range homeworks {
		if len(homeworks[i].GradersJSON) > 0 {
			_ = json.Unmarshal(homeworks[i].GradersJSON, &homeworks[i].Graders)
		}
		homeworks[i].GradersJSON = nil
	}

	return homeworks, totalCount, nil
}
