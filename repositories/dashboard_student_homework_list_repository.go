package repositories

import (
	"fmt"
	"time"

	"be-lms/database/db"
	"be-lms/dto"

	"gorm.io/gorm"
)

type DashboardStudentHomeworkListRepository interface {
	GetStudentHomeworkList(userID int64, courseID *int64, startDate, endDate *int64, limit, offset int) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error)
}

type dashboardStudentHomeworkListRepository struct{}

func NewDashboardStudentHomeworkListRepository() DashboardStudentHomeworkListRepository {
	return &dashboardStudentHomeworkListRepository{}
}

func homeworkUserKey(homeworkID, lessonID int64) string {
	return fmt.Sprintf("%d:%d", homeworkID, lessonID)
}

func (r *dashboardStudentHomeworkListRepository) applyDateFilter(
	query *gorm.DB,
	userID int64,
	startDate, endDate *int64,
) *gorm.DB {
	if startDate == nil || endDate == nil {
		return query
	}
	start := time.Unix(*startDate, 0)
	end := time.Unix(*endDate, 0)
	return query.Where(`(
		(hrl.assigned_at IS NOT NULL AND hrl.assigned_at >= ? AND hrl.assigned_at <= ?)
		OR EXISTS (
			SELECT 1 FROM homework_users hu_filter
			WHERE hu_filter.homework_id = h.id
			  AND hu_filter.user_id = ?
			  AND hu_filter.lesson_id = l.id
			  AND hu_filter.created_at >= ?
			  AND hu_filter.created_at <= ?
		)
	)`, start, end, userID, start, end)
}

func (r *dashboardStudentHomeworkListRepository) GetStudentHomeworkList(
	userID int64,
	courseID *int64,
	startDate, endDate *int64,
	limit, offset int,
) ([]dto.DashboardStudentHomeworkListItemDTO, int64, error) {
	baseQuery := db.ReplicaDB.Table("homework_ref_lessons hrl").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id AND h.deleted_at IS NULL").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID).
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("(hrl.course_id IS NULL OR hrl.course_id = 0 OR hrl.course_id = c.id)")

	if courseID != nil {
		baseQuery = baseQuery.Where("c.id = ?", *courseID)
	}
	baseQuery = r.applyDateFilter(baseQuery, userID, startDate, endDate)

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Select("COUNT(DISTINCT h.id)").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := baseQuery.Session(&gorm.Session{}).
		Select(`
			h.id, h.name, h.description, h.cover_image_info, h.total_questions,
			l.id as lesson_id, l.title as lesson_title,
			c.id as course_id, c.name as course_name,
			hrl.assigned_at as assigned_at
		`).
		Group("h.id, h.name, h.description, h.cover_image_info, h.total_questions, l.id, l.title, c.id, c.name, hrl.assigned_at").
		Order("hrl.assigned_at DESC NULLS LAST, h.id DESC")

	if limit > 0 {
		listQuery = listQuery.Limit(limit)
	}
	if offset > 0 {
		listQuery = listQuery.Offset(offset)
	}

	var homeworksWithMeta []struct {
		ID             int64
		Name           string
		Description    string
		CoverImageInfo string `gorm:"column:cover_image_info"`
		TotalQuestions int64
		LessonID       int64
		LessonTitle    string
		CourseID       int64
		CourseName     string
		AssignedAt     *time.Time
	}
	if err := listQuery.Scan(&homeworksWithMeta).Error; err != nil {
		return nil, 0, err
	}

	homeworkIDs := make([]int64, 0, len(homeworksWithMeta))
	for _, h := range homeworksWithMeta {
		homeworkIDs = append(homeworkIDs, h.ID)
	}

	type HomeworkUserInfo struct {
		HomeworkID         int64
		LessonID           int64
		QuestionsCompleted int64
		CreatedAt          time.Time
		Ratio              float64
	}
	homeworkUserMap := make(map[string]HomeworkUserInfo)
	if len(homeworkIDs) > 0 {
		var homeworkUsers []HomeworkUserInfo
		db.ReplicaDB.Table("homework_users").
			Select("homework_id, lesson_id, questions_completed, created_at, ratio").
			Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
			Scan(&homeworkUsers)
		for _, hu := range homeworkUsers {
			homeworkUserMap[homeworkUserKey(hu.HomeworkID, hu.LessonID)] = hu
		}
	}

	skipQuestionMap := make(map[string]int64)
	if len(homeworkIDs) > 0 {
		type SkipQuestionInfo struct {
			HomeworkID int64
			LessonID   int64
			Count      int64
		}
		var skipQuestions []SkipQuestionInfo
		db.ReplicaDB.Table("homework_user_skip_questions").
			Select("homework_id, lesson_id, COUNT(*) as count").
			Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
			Group("homework_id, lesson_id").
			Scan(&skipQuestions)
		for _, sq := range skipQuestions {
			skipQuestionMap[homeworkUserKey(sq.HomeworkID, sq.LessonID)] = sq.Count
		}
	}

	var result []dto.DashboardStudentHomeworkListItemDTO
	for _, h := range homeworksWithMeta {
		key := homeworkUserKey(h.ID, h.LessonID)
		hu, ok := homeworkUserMap[key]
		isSubmitted := ok && hu.CreatedAt.Unix() > 0
		questionsCompleted := 0
		ratio := 0.0
		createdAt := int64(0)
		if ok {
			createdAt = hu.CreatedAt.Unix()
			ratio = hu.Ratio
			questionsCompleted = int(hu.QuestionsCompleted)
		}
		assignedAt := int64(0)
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
			CourseID:           h.CourseID,
			CourseName:         h.CourseName,
			LessonID:           h.LessonID,
			LessonTitle:        h.LessonTitle,
			Ratio:              ratio,
			SkipQuestionsCount: int(skipQuestionMap[key]),
			AssignedAt:         assignedAt,
		})
	}
	return result, total, nil
}
