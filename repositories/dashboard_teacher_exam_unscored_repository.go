package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/requests"
)

type DashboardTeacherExamUnscoredRepository interface {
	GetUnscoredExams(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherExamUnscoredListRequest) ([]dto.DashboardTeacherExamUnscored, int64, error)
}

type dashboardTeacherExamUnscoredRepository struct{}

func NewDashboardTeacherExamUnscoredRepository() DashboardTeacherExamUnscoredRepository {
	return &dashboardTeacherExamUnscoredRepository{}
}

func (r *dashboardTeacherExamUnscoredRepository) GetUnscoredExams(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherExamUnscoredListRequest) ([]dto.DashboardTeacherExamUnscored, int64, error) {
	var exams []dto.DashboardTeacherExamUnscored
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
			return []dto.DashboardTeacherExamUnscored{}, 0, nil
		}
	}

	query := db.ReplicaDB.Table("exam_users eu").
		Select(`
			eu.user_id, eu.exam_id,
			erl.course_id,
			l.id as lesson_id,
			u.name as student_name,
			c.name as course_name,
			s.name as subject_name,
			e.name as exam_name,
			l.title as lesson_title,
			eu.created_at as submitted_at,
			e.deadline,
			CASE
				WHEN eu.created_at <= e.deadline THEN true
				ELSE false
			END as submit_on_time,
			eu.ratio,
			e.total_questions,
			(SELECT COUNT(*) FROM exam_question_user_manual_scoring equms
			WHERE equms.exam_id = eu.exam_id
			AND equms.user_id = eu.user_id
			AND equms.is_scored = false) as unscored_questions
		`).
		Group("eu.user_id, eu.exam_id, erl.course_id, l.id, u.name, c.name, s.name, e.name, l.title, eu.created_at, e.deadline, eu.ratio, e.total_questions").
		Joins("LEFT JOIN users u ON eu.user_id = u.id").
		Joins("LEFT JOIN exams e ON eu.exam_id = e.id").
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("LEFT JOIN lessons l ON erl.lesson_id = l.id").
		Joins("LEFT JOIN courses c ON erl.course_id = c.id").
		Joins("LEFT JOIN subjects s ON c.subject_id = s.id").
		Where("EXISTS (SELECT 1 FROM exam_question_user_manual_scoring equms WHERE equms.exam_id = eu.exam_id AND equms.user_id = eu.user_id AND equms.is_scored = false)")

	// Nếu chỉ lấy exam cùng khóa (role_id = 2 hoặc 3)
	if onlyUserCourses && userID > 0 {
		query = query.Joins("LEFT JOIN user_courses ON erl.course_id = user_courses.course_id").
			Where("user_courses.user_id = ?", userID)
	}

	// Filter by course_id
	if req.CourseID > 0 {
		query = query.Where("erl.course_id = ?", req.CourseID)
	} else if len(userCourseIDs) > 0 {
		// Không truyền course_id: lọc theo danh sách course_id của user
		query = query.Where("erl.course_id IN (?)", userCourseIDs)
	}

	// Filter by exam_id
	if req.ExamID > 0 {
		query = query.Where("e.id = ?", req.ExamID)
	}

	// Filter by start_date and end_date (only date part, not time)
	if req.StartDate > 0 {
		query = query.Where("DATE(eu.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	}
	if req.EndDate > 0 {
		query = query.Where("DATE(eu.created_at) <= DATE(to_timestamp(?))", req.EndDate)
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
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("eu.created_at DESC")
	}

	err := query.Find(&exams).Error
	return exams, totalCount, err
}

