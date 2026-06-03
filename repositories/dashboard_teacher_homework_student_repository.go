package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
)

type DashboardTeacherHomeworkStudentRepository interface {
	GetStudents(req *requests.DashboardTeacherHomeworkStudentRequest) ([]dto.DashboardTeacherHomeworkStudent, int64, error)
	GetStudentStats(req *requests.DashboardTeacherHomeworkStudentStatsRequest) ([]dto.DashboardTeacherHomeworkStudentStats, int64, error)
	GetHomeworkOverview(req *requests.DashboardTeacherHomeworkOverviewRequest) (*dto.DashboardTeacherHomeworkOverview, error)
}

type dashboardTeacherHomeworkStudentRepository struct{}

func NewDashboardTeacherHomeworkStudentRepository() DashboardTeacherHomeworkStudentRepository {
	return &dashboardTeacherHomeworkStudentRepository{}
}

func (r *dashboardTeacherHomeworkStudentRepository) GetStudents(req *requests.DashboardTeacherHomeworkStudentRequest) ([]dto.DashboardTeacherHomeworkStudent, int64, error) {
	var students []dto.DashboardTeacherHomeworkStudent
	var totalCount int64

	query := db.ReplicaDB.Table("user_courses").
		Select(`
			DISTINCT user_courses.user_id as student_id,
			users.name as student_name,
			users.avatar_info as student_avatar,
			COALESCE(homeworks.total_questions, 0) as total_questions,
			COALESCE(homework_users.questions_completed, 0) as questions_completed
		`).
		Joins("LEFT JOIN users ON user_courses.user_id = users.id").
		Joins("LEFT JOIN courses ON user_courses.course_id = courses.id").
		Joins("LEFT JOIN chapters ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN lessons ON chapters.id = lessons.chapter_id").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id").
		Joins("LEFT JOIN homeworks ON homeworks.id = hrl.homework_id").
		Joins("LEFT JOIN homework_users ON homework_users.homework_id = homeworks.id AND homework_users.user_id = user_courses.user_id").
		Where("users.deleted_at IS NULL").
		Where("homeworks.deleted_at IS NULL").
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("user_courses.course_id = ?", req.CourseID)

	// Add homework_id filter if provided
	if req.HomeworkID > 0 {
		query = query.Where("homeworks.id = ?", req.HomeworkID)
	}

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		query = query.Where("lessons.id = ?", req.LessonID)
	}

	// Count total (distinct by user) - create separate count query
	countQuery := db.ReplicaDB.Table("user_courses").
		Joins("LEFT JOIN users ON user_courses.user_id = users.id").
		Joins("LEFT JOIN courses ON user_courses.course_id = courses.id").
		Joins("LEFT JOIN chapters ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN lessons ON chapters.id = lessons.chapter_id").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id").
		Joins("LEFT JOIN homeworks ON homeworks.id = hrl.homework_id").
		Where("users.deleted_at IS NULL").
		Where("homeworks.deleted_at IS NULL").
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("user_courses.course_id = ?", req.CourseID)

	// Add homework_id filter if provided
	if req.HomeworkID > 0 {
		countQuery = countQuery.Where("homeworks.id = ?", req.HomeworkID)
	}

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		countQuery = countQuery.Where("lessons.id = ?", req.LessonID)
	}

	if err := countQuery.Distinct("user_courses.user_id").Count(&totalCount).Error; err != nil {
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
		query = query.Order("users.name ASC")
	}

	err := query.Find(&students).Error
	return students, totalCount, err
}

func (r *dashboardTeacherHomeworkStudentRepository) GetStudentStats(req *requests.DashboardTeacherHomeworkStudentStatsRequest) ([]dto.DashboardTeacherHomeworkStudentStats, int64, error) {
	var students []dto.DashboardTeacherHomeworkStudentStats
	var totalCount int64

	// 1) Lấy danh sách học sinh (id, name, avatar) thuộc course
	type StudentInfo struct {
		UserID int64  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
		Avatar string `gorm:"column:avatar"`
	}
	var studentInfos []StudentInfo

	// Count total students first (without ORDER BY)
	countQuery := db.ReplicaDB.Table("user_courses uc").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	// Apply lesson filter for count
	if req.LessonID > 0 {
		countQuery = countQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
			Joins("JOIN homeworks h ON h.id = hrl.homework_id").
			Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND l.id = ?", req.LessonID)
	}

	if err := countQuery.Distinct("uc.user_id").Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Build main query for data retrieval
	baseQuery := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, u.name, u.avatar_info as avatar").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	// Apply lesson filter if provided
	if req.LessonID > 0 {
		baseQuery = baseQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
			Joins("JOIN homeworks h ON h.id = hrl.homework_id").
			Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND l.id = ?", req.LessonID)
	}

	// Apply sorting (now u.name is in SELECT list)
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			if field == "users.name" || field == "student_name" || field == "name" {
				baseQuery = baseQuery.Order("u.name " + order)
			}
		}
	} else {
		baseQuery = baseQuery.Order("u.name ASC")
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		baseQuery = baseQuery.Offset(offset).Limit(req.Limit)
	}

	if err := baseQuery.Find(&studentInfos).Error; err != nil {
		return nil, 0, err
	}

	if len(studentInfos) == 0 {
		return []dto.DashboardTeacherHomeworkStudentStats{}, totalCount, nil
	}

	// Extract user IDs for subsequent queries
	userIDs := make([]int64, len(studentInfos))
	for i, s := range studentInfos {
		userIDs[i] = s.UserID
	}

	// Initialize result map
	statsMap := make(map[int64]*dto.DashboardTeacherHomeworkStudentStats)
	for _, s := range studentInfos {
		statsMap[s.UserID] = &dto.DashboardTeacherHomeworkStudentStats{
			StudentID:   s.UserID,
			StudentName: s.Name,
			// Avatar will be handled in resource layer
		}
	}

	// 2) total_homeworks (join với user_courses join homework_ref_lessons) where course_id, lesson_id
	var totalHomeworkRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	totalHomeworkQuery := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, COUNT(DISTINCT h.id) as count").
		Joins("JOIN courses c ON c.id = uc.course_id").
		Joins("JOIN chapters ch ON ch.program_id = c.program_id").
		Joins("JOIN lessons l ON l.chapter_id = ch.id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("uc.user_id IN (?)", userIDs)
	
	if req.LessonID > 0 {
		totalHomeworkQuery = totalHomeworkQuery.Where("l.id = ?", req.LessonID)
	}
	
	if err := totalHomeworkQuery.Group("uc.user_id").Find(&totalHomeworkRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range totalHomeworkRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.TotalHomework = row.Count
		}
	}

	// 3) total_assigned_homeworks (where hrl.assigned_at != null)
	var totalAssignedRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	totalAssignedQuery := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, COUNT(DISTINCT h.id) as count").
		Joins("JOIN courses c ON c.id = uc.course_id").
		Joins("JOIN chapters ch ON ch.program_id = c.program_id").
		Joins("JOIN lessons l ON l.chapter_id = ch.id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND hrl.assigned_at IS NOT NULL AND uc.course_id = ?", req.CourseID).
		Where("uc.user_id IN (?)", userIDs)
	
	if req.LessonID > 0 {
		totalAssignedQuery = totalAssignedQuery.Where("l.id = ?", req.LessonID)
	}
	
	if err := totalAssignedQuery.Group("uc.user_id").Find(&totalAssignedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range totalAssignedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.TotalAssignedHomeworks = row.Count
		}
	}

	// 4) in_progress_homework (join với homework_users join với user_courses join homework_ref_lessons join homeworks) 
	// where homework_users.questions_completed < homeworks.total_questions
	var inProgressRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	inProgressQuery := db.ReplicaDB.Table("homework_users hu").
		Select("hu.user_id, COUNT(DISTINCT hu.homework_id) as count").
		Joins("JOIN homeworks h ON h.id = hu.homework_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = hu.user_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.questions_completed < h.total_questions")
	
	if req.LessonID > 0 {
		inProgressQuery = inProgressQuery.Where("l.id = ?", req.LessonID)
	}
	
	if err := inProgressQuery.Group("hu.user_id").Find(&inProgressRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range inProgressRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.InProgressHomework = row.Count
		}
	}

	// 5) completed_homework (where homework_users.questions_completed >= homeworks.total_questions)
	var completedRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	completedQuery := db.ReplicaDB.Table("homework_users hu").
		Select("hu.user_id, COUNT(DISTINCT hu.homework_id) as count").
		Joins("JOIN homeworks h ON h.id = hu.homework_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = hu.user_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.questions_completed >= h.total_questions")
	
	if req.LessonID > 0 {
		completedQuery = completedQuery.Where("l.id = ?", req.LessonID)
	}
	
	if err := completedQuery.Group("hu.user_id").Find(&completedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range completedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.CompletedHomework = row.Count
		}
	}

	// 6) average_ratio (lấy trung bình hu.ratio)
	var avgRatioRows []struct {
		UserID int64   `gorm:"column:user_id"`
		Avg    float64 `gorm:"column:avg_ratio"`
	}
	avgRatioQuery := db.ReplicaDB.Table("homework_users hu").
		Select("hu.user_id, AVG(hu.ratio) as avg_ratio").
		Joins("JOIN homeworks h ON h.id = hu.homework_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = hu.user_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.ratio IS NOT NULL")
	
	if req.LessonID > 0 {
		avgRatioQuery = avgRatioQuery.Where("l.id = ?", req.LessonID)
	}
	
	if err := avgRatioQuery.Group("hu.user_id").Find(&avgRatioRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range avgRatioRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.AverageRatio = row.Avg
		}
	}

	// 7) Tính not_started_homework = total_assigned_homeworks - in_progress_homework - completed_homework
	for _, stats := range statsMap {
		stats.NotStartedHomework = stats.TotalAssignedHomeworks - stats.InProgressHomework - stats.CompletedHomework
	}

	// 8) Convert map to slice maintaining order
	students = make([]dto.DashboardTeacherHomeworkStudentStats, 0, len(studentInfos))
	for _, studentInfo := range studentInfos {
		if stats, ok := statsMap[studentInfo.UserID]; ok {
			students = append(students, *stats)
		}
	}

	return students, totalCount, nil
}

func (r *dashboardTeacherHomeworkStudentRepository) GetHomeworkOverview(req *requests.DashboardTeacherHomeworkOverviewRequest) (*dto.DashboardTeacherHomeworkOverview, error) {
	var overview dto.DashboardTeacherHomeworkOverview

	// Query để lấy tổng số học sinh trong course của homework này
	var totalStudents int64
	query := db.ReplicaDB.Table("user_courses").
		Joins("JOIN courses ON user_courses.course_id = courses.id").
		Joins("JOIN chapters ON courses.program_id = chapters.program_id").
		Joins("JOIN lessons ON chapters.id = lessons.chapter_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id").
		Joins("JOIN homeworks ON hrl.homework_id = homeworks.id").
		Joins("JOIN users ON user_courses.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", req.HomeworkID)

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		query = query.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		query = query.Where("user_courses.course_id = ?", req.CourseID)
	}

	err := query.Count(&totalStudents).Error
	if err != nil {
		return nil, err
	}

	overview.TotalStudents = totalStudents

	// Query để lấy số học sinh đang làm (questions_completed < total_questions)
	var inProgressCount int64
	inProgressQuery := db.ReplicaDB.Table("homework_users").
		Joins("JOIN homeworks ON homework_users.homework_id = homeworks.id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
		Joins("JOIN lessons ON hrl.lesson_id = lessons.id").
		Joins("JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("JOIN courses ON chapters.program_id = courses.program_id").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id AND user_courses.user_id = homework_users.user_id").
		Joins("JOIN users ON homework_users.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL AND homework_users.questions_completed < homeworks.total_questions", req.HomeworkID)

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		inProgressQuery = inProgressQuery.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		inProgressQuery = inProgressQuery.Where("user_courses.course_id = ?", req.CourseID)
	}

	err = inProgressQuery.Count(&inProgressCount).Error
	if err != nil {
		return nil, err
	}

	overview.InProgressCount = inProgressCount

	// Query để lấy số học sinh đã làm xong (questions_completed >= total_questions)
	var completedCount int64
	completedQuery := db.ReplicaDB.Table("homework_users").
		Joins("JOIN homeworks ON homework_users.homework_id = homeworks.id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
		Joins("JOIN lessons ON hrl.lesson_id = lessons.id").
		Joins("JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("JOIN courses ON chapters.program_id = courses.program_id").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id AND user_courses.user_id = homework_users.user_id").
		Joins("JOIN users ON homework_users.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL AND homework_users.questions_completed >= homeworks.total_questions", req.HomeworkID)

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		completedQuery = completedQuery.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		completedQuery = completedQuery.Where("user_courses.course_id = ?", req.CourseID)
	}

	err = completedQuery.Count(&completedCount).Error
	if err != nil {
		return nil, err
	}

	overview.CompletedCount = completedCount

	// Tính số học sinh chưa làm = tổng - đã làm xong - đang làm
	overview.NotStartedCount = totalStudents - completedCount - inProgressCount

	// Tính phần trăm
	if totalStudents > 0 {
		overview.NotStartedPercentage = float64(overview.NotStartedCount) / float64(totalStudents) * 100
		overview.InProgressPercentage = float64(inProgressCount) / float64(totalStudents) * 100
		overview.CompletedPercentage = float64(completedCount) / float64(totalStudents) * 100
	}

	return &overview, nil
}
