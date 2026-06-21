package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
)

type DashboardTeacherExerciseStudentRepository interface {
	GetStudentStats(req *requests.DashboardTeacherExerciseStudentStatsRequest) ([]dto.DashboardTeacherExerciseStudentStats, int64, error)
}

type dashboardTeacherExerciseStudentRepository struct{}

func NewDashboardTeacherExerciseStudentRepository() DashboardTeacherExerciseStudentRepository {
	return &dashboardTeacherExerciseStudentRepository{}
}


func (r *dashboardTeacherExerciseStudentRepository) GetStudentStats(req *requests.DashboardTeacherExerciseStudentStatsRequest) ([]dto.DashboardTeacherExerciseStudentStats, int64, error) {
	type StudentInfo struct {
		UserID int64  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
	}
	var studentInfos []StudentInfo
	var totalCount int64

	countQuery := db.ReplicaDB.Table("user_courses uc").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	if req.LessonID > 0 {
		countQuery = countQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN exercise_ref_lessons erl ON erl.lesson_id = l.id").
			Joins("JOIN exercises e ON e.id = erl.exercise_id").
			Where("e.deleted_at IS NULL AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND (erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = uc.course_id) AND l.id = ?", req.LessonID)
	}

	if err := countQuery.Distinct("uc.user_id").Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	baseQuery := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	if req.LessonID > 0 {
		baseQuery = baseQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN exercise_ref_lessons erl ON erl.lesson_id = l.id").
			Joins("JOIN exercises e ON e.id = erl.exercise_id").
			Where("e.deleted_at IS NULL AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND (erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = uc.course_id) AND l.id = ?", req.LessonID)
	}

	baseQuery = baseQuery.Order("u.name ASC")
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		baseQuery = baseQuery.Offset(offset).Limit(req.Limit)
	}

	if err := baseQuery.Find(&studentInfos).Error; err != nil {
		return nil, 0, err
	}
	if len(studentInfos) == 0 {
		return []dto.DashboardTeacherExerciseStudentStats{}, totalCount, nil
	}

	userIDs := make([]int64, len(studentInfos))
	statsMap := make(map[int64]*dto.DashboardTeacherExerciseStudentStats, len(studentInfos))
	for i, s := range studentInfos {
		userIDs[i] = s.UserID
		statsMap[s.UserID] = &dto.DashboardTeacherExerciseStudentStats{
			StudentID:   s.UserID,
			StudentName: s.Name,
		}
	}

	type countRow struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}

	assignedFilter := "e.deleted_at IS NULL AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND (erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = uc.course_id) AND uc.course_id = ?"

	// total assigned exercises
	var assignedRows []countRow
	assignedQ := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, COUNT(DISTINCT e.id) as count").
		Joins("JOIN courses c ON c.id = uc.course_id").
		Joins("JOIN chapters ch ON ch.program_id = c.program_id").
		Joins("JOIN lessons l ON l.chapter_id = ch.id").
		Joins("JOIN exercise_ref_lessons erl ON erl.lesson_id = l.id").
		Joins("JOIN exercises e ON e.id = erl.exercise_id").
		Where(assignedFilter, req.CourseID).
		Where("uc.user_id IN ?", userIDs)
	if req.LessonID > 0 {
		assignedQ = assignedQ.Where("l.id = ?", req.LessonID)
	}
	if err := assignedQ.Group("uc.user_id").Find(&assignedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range assignedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.TotalExercises = row.Count
			stats.TotalAssignedExercises = row.Count
		}
	}

	// in progress: có câu trả lời nhưng chưa có bản ghi exercise_users (chưa nộp)
	var inProgressRows []countRow
	inProgressQ := db.ReplicaDB.Table("exercise_question_users equ").
		Select("equ.user_id, COUNT(DISTINCT equ.exercise_id) as count").
		Joins("JOIN exercises e ON e.id = equ.exercise_id AND e.deleted_at IS NULL").
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id").
		Joins("JOIN lessons l ON l.id = erl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL AND c.id = ?", req.CourseID).
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = equ.user_id").
		Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Where("(erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = uc.course_id)").
		Where("equ.user_id IN ?", userIDs).
		Where(`NOT EXISTS (
			SELECT 1 FROM exercise_users eu
			WHERE eu.exercise_id = e.id AND eu.user_id = equ.user_id
		)`)
	if req.LessonID > 0 {
		inProgressQ = inProgressQ.Where("l.id = ?", req.LessonID)
	}
	if err := inProgressQ.Group("equ.user_id").Find(&inProgressRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range inProgressRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.InProgressExercises = row.Count
		}
	}

	// completed: đã nộp (có exercise_users) cho bài được giao trong khóa
	var completedRows []countRow
	completedQ := db.ReplicaDB.Table("exercise_users eu").
		Select("eu.user_id, COUNT(DISTINCT eu.exercise_id) as count").
		Joins("JOIN exercises e ON e.id = eu.exercise_id AND e.deleted_at IS NULL").
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id").
		Joins("JOIN lessons l ON l.id = erl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL AND c.id = ?", req.CourseID).
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = eu.user_id").
		Where(assignedFilter, req.CourseID).
		Where("eu.user_id IN ?", userIDs)
	if req.LessonID > 0 {
		completedQ = completedQ.Where("l.id = ?", req.LessonID)
	}
	if err := completedQ.Group("eu.user_id").Find(&completedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range completedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.CompletedExercises = row.Count
		}
	}

	// average ratio
	type avgRow struct {
		UserID int64   `gorm:"column:user_id"`
		Avg    float64 `gorm:"column:avg_ratio"`
	}
	var avgRows []avgRow
	avgQ := db.ReplicaDB.Table("exercise_users eu").
		Select("eu.user_id, AVG(eu.ratio) as avg_ratio").
		Joins("JOIN exercises e ON e.id = eu.exercise_id AND e.deleted_at IS NULL").
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id").
		Joins("JOIN lessons l ON l.id = erl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL AND c.id = ?", req.CourseID).
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = eu.user_id").
		Where(assignedFilter, req.CourseID).
		Where("eu.user_id IN ?", userIDs).
		Where("eu.ratio IS NOT NULL")
	if req.LessonID > 0 {
		avgQ = avgQ.Where("l.id = ?", req.LessonID)
	}
	if err := avgQ.Group("eu.user_id").Find(&avgRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range avgRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.AverageRatio = row.Avg
		}
	}

	students := make([]dto.DashboardTeacherExerciseStudentStats, 0, len(studentInfos))
	for _, info := range studentInfos {
		if stats, ok := statsMap[info.UserID]; ok {
			stats.NotStartedExercises = stats.TotalAssignedExercises - stats.InProgressExercises - stats.CompletedExercises
			if stats.NotStartedExercises < 0 {
				stats.NotStartedExercises = 0
			}
			students = append(students, *stats)
		}
	}

	return students, totalCount, nil
}
