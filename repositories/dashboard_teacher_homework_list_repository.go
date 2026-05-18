package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/requests"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type DashboardTeacherHomeworkListRepository interface {
	GetHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) ([]dto.DashboardTeacherHomeworkListItem, int64, error)
}

type dashboardTeacherHomeworkListRepository struct{}

func NewDashboardTeacherHomeworkListRepository() DashboardTeacherHomeworkListRepository {
	return &dashboardTeacherHomeworkListRepository{}
}

func (r *dashboardTeacherHomeworkListRepository) GetHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) ([]dto.DashboardTeacherHomeworkListItem, int64, error) {
	var homeworks []dto.DashboardTeacherHomeworkListItem
	var totalCount int64

	var _, endDate *time.Time
	if req.StartDate != "" {
		if timestamp, err := strconv.ParseInt(req.StartDate, 10, 64); err == nil {
			t := time.Unix(timestamp, 0)
			_ = &t
		} else if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			_ = &t
		}
	}
	if req.EndDate != "" {
		if timestamp, err := strconv.ParseInt(req.EndDate, 10, 64); err == nil {
			t := time.Unix(timestamp, 0)
			endDate = &t
		} else if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = &t
		}
	}

	var homeworkIDs []int64
	if req.HomeworkIDs != "" {
		idsStr := strings.Split(req.HomeworkIDs, ",")
		for _, idStr := range idsStr {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				homeworkIDs = append(homeworkIDs, id)
			}
		}
	}

	var lessonIDs []int64
	if req.LessonIDs != "" {
		idsStr := strings.Split(req.LessonIDs, ",")
		for _, idStr := range idsStr {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				lessonIDs = append(lessonIDs, id)
			}
		}
	}
	if req.LessonID > 0 && len(lessonIDs) == 0 {
		lessonIDs = append(lessonIDs, req.LessonID)
	}

	baseQuery := db.ReplicaDB.Table("homeworks h").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.id = hrl.course_id").
		Joins("JOIN lesson_schedules ls ON ls.course_id = hrl.course_id AND ls.lesson_id = hrl.lesson_id").
		Joins("JOIN weeks w ON w.id = ls.week_id").
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_at IS NOT NULL").
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("hrl.course_id = ?", req.CourseID).
		Where("c.id = ?", req.CourseID)

	if req.StudentID > 0 {
		baseQuery = baseQuery.Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ?", req.StudentID)
	} else {
		baseQuery = baseQuery.Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id IS NULL")
	}

	if req.ChapterID > 0 {
		baseQuery = baseQuery.Where("ch.id = ?", req.ChapterID)
	}

	if len(lessonIDs) > 0 {
		baseQuery = baseQuery.Where("l.id IN (?)", lessonIDs)
	}

	if len(homeworkIDs) > 0 {
		baseQuery = baseQuery.Where("h.id IN (?)", homeworkIDs)
	}

	if endDate != nil {
		baseQuery = baseQuery.Where("hrl.created_at <= ?", *endDate)
	}

	// Bỏ logic lọc theo week.start_date và week.end_date
	// if startDate != nil && endDate != nil {
	// 	baseQuery = baseQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	// }

	countQuery := baseQuery.Session(&gorm.Session{}).Distinct("h.id")
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	selectFields := `
			h.id as homework_id,
			h.name as homework_name,
			h.total_questions,
			COALESCE(MAX(hu.questions_completed), 0) as questions_completed,
			CASE 
				WHEN MAX(hu.questions_completed) IS NULL THEN 'not_started'
				WHEN MAX(hu.questions_completed) >= h.total_questions THEN 'completed'
				ELSE 'in_progress'
			END as status,
			COALESCE(MAX(EXTRACT(EPOCH FROM hu.updated_at)::bigint), 0) as submitted_at,
			COALESCE(MAX(EXTRACT(EPOCH FROM hrl.assigned_at)::bigint), 0) as assigned_at,
			MIN(l.id) as lesson_id,
			MIN(l.title) as lesson_name`

	var dataQuery *gorm.DB
	if endDate != nil {
		selectFields += `,
			CASE 
				WHEN MAX(hrl.assigned_at) > ? THEN TRUE
				ELSE FALSE
			END as assign_late
		`
		dataQuery = baseQuery.Session(&gorm.Session{}).Select(selectFields, *endDate)
	} else {
		selectFields += `,
			FALSE as assign_late
		`
		dataQuery = baseQuery.Session(&gorm.Session{}).Select(selectFields)
	}

	dataQuery = dataQuery.Group("h.id, h.name, h.total_questions")

	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			dataQuery = dataQuery.Order(field + " " + order)
		}
	} else {
		dataQuery = dataQuery.Order("lesson_id ASC, homework_id ASC")
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		dataQuery = dataQuery.Offset(offset).Limit(req.Limit)
	}

	if err := dataQuery.Find(&homeworks).Error; err != nil {
		return nil, 0, err
	}

	return homeworks, totalCount, nil
}
