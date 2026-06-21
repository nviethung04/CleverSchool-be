package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type DashboardTeacherHomeworkListRepository interface {
	GetStudentHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) ([]dto.DashboardTeacherHomeworkListItem, int64, error)
}

type dashboardTeacherHomeworkListRepository struct{}

func NewDashboardTeacherHomeworkListRepository() DashboardTeacherHomeworkListRepository {
	return &dashboardTeacherHomeworkListRepository{}
}

func parseCSVInt64s(raw string) []int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.ParseInt(part, 10, 64); err == nil && id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func (r *dashboardTeacherHomeworkListRepository) applyHomeworkListFilters(query *gorm.DB, req *requests.DashboardTeacherHomeworkListRequest) *gorm.DB {
	if req.ChapterID > 0 {
		query = query.Where("ch.id = ?", req.ChapterID)
	}
	if lessonIDs := parseCSVInt64s(req.LessonIDs); len(lessonIDs) > 0 {
		query = query.Where("l.id IN ?", lessonIDs)
	}
	if homeworkIDs := parseCSVInt64s(req.HomeworkIDs); len(homeworkIDs) > 0 {
		query = query.Where("h.id IN ?", homeworkIDs)
	}
	if req.StartDate > 0 && req.EndDate > 0 {
		query = query.Where(
			"hrl.assigned_at IS NOT NULL AND DATE(hrl.assigned_at) >= DATE(to_timestamp(?)) AND DATE(hrl.assigned_at) <= DATE(to_timestamp(?))",
			req.StartDate, req.EndDate,
		)
	} else if req.StartDate > 0 {
		query = query.Where("hrl.assigned_at IS NOT NULL AND DATE(hrl.assigned_at) >= DATE(to_timestamp(?))", req.StartDate)
	} else if req.EndDate > 0 {
		query = query.Where("hrl.assigned_at IS NOT NULL AND DATE(hrl.assigned_at) <= DATE(to_timestamp(?))", req.EndDate)
	}
	return query
}

func (r *dashboardTeacherHomeworkListRepository) buildHomeworkListQuery(req *requests.DashboardTeacherHomeworkListRequest) *gorm.DB {
	q := db.ReplicaDB.Table("homework_ref_lessons hrl").
		Select(`
			h.id as homework_id,
			h.name as homework_name,
			COALESCE(h.total_questions, 0) as total_questions,
			COALESCE(hu.questions_completed, 0) as questions_completed,
			CASE WHEN hu.id IS NULL THEN false ELSE true END as has_submission,
			COALESCE(CAST(extract(epoch from hu.created_at) AS BIGINT), 0) as submitted_at,
			l.id as lesson_id,
			l.title as lesson_name,
			COALESCE(CAST(extract(epoch from hrl.assigned_at) AS BIGINT), 0) as assigned_at,
			false as assign_late
		`).
		Joins("JOIN homeworks h ON h.id = hrl.homework_id AND h.deleted_at IS NULL").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL AND c.id = ?", req.CourseID).
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", req.StudentID).
		Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = uc.user_id AND hu.lesson_id = l.id").
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("(hrl.course_id IS NULL OR hrl.course_id = 0 OR hrl.course_id = ?)", req.CourseID)

	return r.applyHomeworkListFilters(q, req)
}

func (r *dashboardTeacherHomeworkListRepository) GetStudentHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) ([]dto.DashboardTeacherHomeworkListItem, int64, error) {
	base := r.buildHomeworkListQuery(req)

	var total int64
	countQuery := base.Session(&gorm.Session{}).Select("COUNT(DISTINCT h.id)")
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := base.Session(&gorm.Session{})
	switch req.OrderBy {
	case "average_score":
		query = query.Order("hu.ratio DESC NULLS LAST, h.name ASC")
	default:
		query = query.Order("hu.questions_completed DESC NULLS LAST, h.name ASC")
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	var rows []dto.DashboardTeacherHomeworkListItem
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
