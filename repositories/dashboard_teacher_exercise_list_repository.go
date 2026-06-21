package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"

	"gorm.io/gorm"
)

type DashboardTeacherExerciseListRepository interface {
	GetStudentExerciseList(req *requests.DashboardTeacherExerciseListRequest) ([]dto.DashboardTeacherExerciseListItem, int64, error)
}

type dashboardTeacherExerciseListRepository struct{}

func NewDashboardTeacherExerciseListRepository() DashboardTeacherExerciseListRepository {
	return &dashboardTeacherExerciseListRepository{}
}

func parseExerciseCSVInt64s(raw string) []int64 {
	return parseCSVInt64s(raw)
}

func (r *dashboardTeacherExerciseListRepository) applyExerciseListFilters(query *gorm.DB, req *requests.DashboardTeacherExerciseListRequest) *gorm.DB {
	if req.ChapterID > 0 {
		query = query.Where("ch.id = ?", req.ChapterID)
	}
	if lessonIDs := parseExerciseCSVInt64s(req.LessonIDs); len(lessonIDs) > 0 {
		query = query.Where("l.id IN ?", lessonIDs)
	}
	if exerciseIDs := parseExerciseCSVInt64s(req.ExerciseIDs); len(exerciseIDs) > 0 {
		query = query.Where("e.id IN ?", exerciseIDs)
	}
	if req.StartDate > 0 && req.EndDate > 0 {
		query = query.Where(
			"erl.assigned_at IS NOT NULL AND DATE(erl.assigned_at) >= DATE(to_timestamp(?)) AND DATE(erl.assigned_at) <= DATE(to_timestamp(?))",
			req.StartDate, req.EndDate,
		)
	} else if req.StartDate > 0 {
		query = query.Where("erl.assigned_at IS NOT NULL AND DATE(erl.assigned_at) >= DATE(to_timestamp(?))", req.StartDate)
	} else if req.EndDate > 0 {
		query = query.Where("erl.assigned_at IS NOT NULL AND DATE(erl.assigned_at) <= DATE(to_timestamp(?))", req.EndDate)
	}
	return query
}

func (r *dashboardTeacherExerciseListRepository) buildExerciseListQuery(req *requests.DashboardTeacherExerciseListRequest) *gorm.DB {
	qcSub := `(SELECT COUNT(DISTINCT equ.question_id) FROM exercise_question_users equ WHERE equ.exercise_id = e.id AND equ.user_id = uc.user_id)`

	q := db.ReplicaDB.Table("exercise_ref_lessons erl").
		Select(`
			e.id as exercise_id,
			e.name as exercise_name,
			COALESCE(e.total_questions, 0) as total_questions,
			COALESCE(`+qcSub+`, 0) as questions_completed,
			CASE WHEN eu.id IS NULL THEN false ELSE true END as has_submission,
			COALESCE(CAST(extract(epoch from eu.created_at) AS BIGINT), 0) as submitted_at,
			COALESCE(eu.ratio, 0) as ratio,
			l.id as lesson_id,
			l.title as lesson_name,
			COALESCE(CAST(extract(epoch from erl.assigned_at) AS BIGINT), 0) as assigned_at,
			false as assign_late
		`).
		Joins("JOIN exercises e ON e.id = erl.exercise_id AND e.deleted_at IS NULL").
		Joins("JOIN lessons l ON l.id = erl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL AND c.id = ?", req.CourseID).
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", req.StudentID).
		Joins("LEFT JOIN exercise_users eu ON eu.exercise_id = e.id AND eu.user_id = uc.user_id").
		Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Where("(erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = ?)", req.CourseID)

	return r.applyExerciseListFilters(q, req)
}

func (r *dashboardTeacherExerciseListRepository) GetStudentExerciseList(req *requests.DashboardTeacherExerciseListRequest) ([]dto.DashboardTeacherExerciseListItem, int64, error) {
	base := r.buildExerciseListQuery(req)

	var total int64
	countQuery := base.Session(&gorm.Session{}).Select("COUNT(DISTINCT e.id)")
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := base.Session(&gorm.Session{})
	switch req.OrderBy {
	case "average_score":
		query = query.Order("eu.ratio DESC NULLS LAST, e.name ASC")
	default:
		query = query.Order("questions_completed DESC NULLS LAST, e.name ASC")
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	var rows []dto.DashboardTeacherExerciseListItem
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
