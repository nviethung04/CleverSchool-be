package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/requests"
	"sort"
	"strconv"
	"strings"
	"time"
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
			user_courses.user_id as student_id,
			users.name as student_name,
			users.avatar_info as student_avatar,
			COALESCE(MAX(homeworks.total_questions), 0) as total_questions,
			COALESCE(MAX(homework_users.questions_completed), 0) as questions_completed
		`).
		Joins("JOIN users ON user_courses.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN courses ON user_courses.course_id = courses.id").
		Joins("JOIN chapters ON courses.program_id = chapters.program_id").
		Joins("JOIN lessons ON chapters.id = lessons.chapter_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id AND hrl.course_id = user_courses.course_id").
		Joins("JOIN homeworks ON homeworks.id = hrl.homework_id").
		Joins("LEFT JOIN homework_users ON homework_users.homework_id = homeworks.id AND homework_users.user_id = user_courses.user_id").
		Where("users.deleted_at IS NULL").
		Where("urr.role_id = 3").
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

	// Apply GROUP BY to ensure unique user_id (always needed to avoid duplicates)
	query = query.Group("user_courses.user_id, users.name, users.avatar_info")

	// Count total (distinct by user) - create separate count query
	countQuery := db.ReplicaDB.Table("user_courses").
		Joins("JOIN users ON user_courses.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN courses ON user_courses.course_id = courses.id").
		Joins("JOIN chapters ON courses.program_id = chapters.program_id").
		Joins("JOIN lessons ON chapters.id = lessons.chapter_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id AND hrl.course_id = user_courses.course_id").
		Joins("JOIN homeworks ON homeworks.id = hrl.homework_id").
		Where("users.deleted_at IS NULL").
		Where("urr.role_id = 3").
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
	if err != nil {
		return nil, 0, err
	}

	if len(students) == 0 {
		return []dto.DashboardTeacherHomeworkStudent{}, totalCount, nil
	}

	studentIDs := make([]int64, len(students))
	for i, s := range students {
		studentIDs[i] = s.StudentID
	}

	homeworkTotalsBase := db.ReplicaDB.Table("homeworks h").
		Select(`
			uc.user_id as student_id,
			h.id as homework_id,
			MAX(h.total_questions) as total_questions,
			MAX(COALESCE(hu.questions_completed, 0)) as questions_completed
		`).
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id").
		Joins("JOIN users u ON uc.user_id = u.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = uc.user_id").
		Where("u.deleted_at IS NULL").
		Where("urr.role_id = 3").
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0").
		Where("uc.course_id = ?", req.CourseID).
		Where("uc.user_id IN (?)", studentIDs)

	if req.HomeworkID > 0 {
		homeworkTotalsBase = homeworkTotalsBase.Where("h.id = ?", req.HomeworkID)
	}
	if req.LessonID > 0 {
		homeworkTotalsBase = homeworkTotalsBase.Where("l.id = ?", req.LessonID)
	}

	homeworkTotalsBase = homeworkTotalsBase.Group("uc.user_id, h.id")

	homeworkTotalsQuery := db.ReplicaDB.Table("(?) as homework_totals", homeworkTotalsBase).
		Select("student_id, SUM(total_questions) as total_questions, SUM(questions_completed) as questions_completed").
		Group("student_id")

	type homeworkTotals struct {
		StudentID          int64
		TotalQuestions     int64
		QuestionsCompleted int64
	}

	var totals []homeworkTotals
	if err := homeworkTotalsQuery.Find(&totals).Error; err != nil {
		return nil, 0, err
	}

	totalMap := make(map[int64]homeworkTotals)
	for _, t := range totals {
		totalMap[t.StudentID] = t
	}

	for i, student := range students {
		if totals, ok := totalMap[student.StudentID]; ok {
			students[i].TotalQuestions = totals.TotalQuestions
			students[i].QuestionsCompleted = totals.QuestionsCompleted
		} else {
			students[i].TotalQuestions = 0
			students[i].QuestionsCompleted = 0
		}
	}

	return students, totalCount, nil
}

func (r *dashboardTeacherHomeworkStudentRepository) GetStudentStats(req *requests.DashboardTeacherHomeworkStudentStatsRequest) ([]dto.DashboardTeacherHomeworkStudentStats, int64, error) {
	var students []dto.DashboardTeacherHomeworkStudentStats
	var totalCount int64

	// Parse start_date và end_date (có thể là Unix timestamp hoặc YYYY-MM-DD)
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if timestamp, err := strconv.ParseInt(req.StartDate, 10, 64); err == nil {
			// Unix timestamp
			t := time.Unix(timestamp, 0)
			startDate = &t
		} else if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			// YYYY-MM-DD format
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if timestamp, err := strconv.ParseInt(req.EndDate, 10, 64); err == nil {
			// Unix timestamp
			t := time.Unix(timestamp, 0)
			endDate = &t
		} else if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			// YYYY-MM-DD format
			endDate = &t
		}
	}

	// Parse homework_ids từ string (cách nhau bởi dấu phẩy) thành array
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

	// Parse lesson_ids từ string (cách nhau bởi dấu phẩy) thành array
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
	// Nếu có LessonID (single) và chưa có LessonIDs, thêm vào
	if req.LessonID > 0 && len(lessonIDs) == 0 {
		lessonIDs = append(lessonIDs, req.LessonID)
	}

	// 1) Lấy danh sách học sinh (id, name, avatar) thuộc course
	type StudentInfo struct {
		UserID int64  `gorm:"column:user_id"`
		Name   string `gorm:"column:name"`
		Avatar string `gorm:"column:avatar"`
	}
	var studentInfos []StudentInfo

	// Count total students first (lấy tất cả học sinh trong course, không filter theo chapter_id/lesson_id)
	countQuery := db.ReplicaDB.Table("user_courses uc").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	// Chỉ filter học sinh theo lesson_ids nếu có (không filter theo chapter_id để hiển thị đầy đủ học sinh)
	if len(lessonIDs) > 0 {
		countQuery = countQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
			Joins("JOIN homeworks h ON h.id = hrl.homework_id").
			Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND l.id IN (?)", lessonIDs)
		countQuery = countQuery.Where("hrl.course_id = uc.course_id")
	}

	if err := countQuery.Distinct("uc.user_id").Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Build main query for data retrieval (lấy tất cả học sinh trong course, không filter theo chapter_id)
	baseQuery := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id, u.name, u.avatar_info as avatar").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", req.CourseID)

	// Chỉ filter học sinh theo lesson_ids nếu có (không filter theo chapter_id để hiển thị đầy đủ học sinh)
	if len(lessonIDs) > 0 {
		baseQuery = baseQuery.
			Joins("JOIN courses c ON c.id = uc.course_id").
			Joins("JOIN chapters ch ON ch.program_id = c.program_id").
			Joins("JOIN lessons l ON l.chapter_id = ch.id").
			Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id").
			Joins("JOIN homeworks h ON h.id = hrl.homework_id").
			Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND l.id IN (?)", lessonIDs)
		baseQuery = baseQuery.Where("hrl.course_id = uc.course_id")
	}

	// Apply GROUP BY to ensure unique user_id (works better with ORDER BY than DISTINCT)
	baseQuery = baseQuery.Group("uc.user_id, u.name, u.avatar_info")

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

	// 2) total_homeworks (join với user_courses join homework_ref_lessons join lesson_schedules join weeks) where course_id, lesson_id
	// Logic: homework → lesson → chapter → program → course (không dựa vào hrl.course_id)
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
		Joins("JOIN lesson_schedules ls ON ls.lesson_id = l.id AND ls.course_id = c.id").
		Joins("JOIN weeks w ON w.id = ls.week_id").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id").
		Where("h.deleted_at IS NULL AND uc.course_id = ?", req.CourseID).
		Where("uc.user_id IN (?)", userIDs)

	// Filter theo start_date và end_date nếu có
	if startDate != nil && endDate != nil {
		totalHomeworkQuery = totalHomeworkQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}

	if len(lessonIDs) > 0 {
		totalHomeworkQuery = totalHomeworkQuery.Where("l.id IN (?)", lessonIDs)
	}

	if req.ChapterID > 0 {
		totalHomeworkQuery = totalHomeworkQuery.Where("ch.id = ?", req.ChapterID)
	}

	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		totalHomeworkQuery = totalHomeworkQuery.Where("h.id IN (?)", homeworkIDs)
	}

	if err := totalHomeworkQuery.Group("uc.user_id").Find(&totalHomeworkRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range totalHomeworkRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.TotalHomework = row.Count
		}
	}

	// 3) total_assigned_homeworks (where hrl.assigned_at != null) join với lesson_schedules và weeks
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
		Joins("JOIN lesson_schedules ls ON ls.course_id = c.id AND ls.lesson_id = l.id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN homeworks h ON h.id = hrl.homework_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_at IS NOT NULL AND uc.course_id = ?", req.CourseID).
		Where("uc.user_id IN (?)", userIDs)

	if len(lessonIDs) > 0 {
		totalAssignedQuery = totalAssignedQuery.Where("l.id IN (?)", lessonIDs)
	}

	if req.ChapterID > 0 {
		totalAssignedQuery = totalAssignedQuery.Where("ch.id = ?", req.ChapterID)
	}

	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		totalAssignedQuery = totalAssignedQuery.Where("h.id IN (?)", homeworkIDs)
	}

	// Chuẩn hoá theo logic list-homeworks: chỉ tính những homework gán đúng course hiện tại
	totalAssignedQuery = totalAssignedQuery.Where("hrl.course_id = uc.course_id")

	// Filter theo start_date và end_date (đồng bộ với list-homeworks)
	if endDate != nil {
		totalAssignedQuery = totalAssignedQuery.Where("hrl.created_at <= ?", *endDate)
	}

	if startDate != nil && endDate != nil {
		totalAssignedQuery = totalAssignedQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}

	if err := totalAssignedQuery.Group("uc.user_id").Find(&totalAssignedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range totalAssignedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.TotalAssignedHomeworks = row.Count
		}
	}

	// 4) in_progress_homework (join với homework_users join với user_courses join homework_ref_lessons join lesson_schedules join weeks join homeworks)
	// where homework_users.questions_completed < homeworks.total_questions
	var inProgressRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	inProgressQuery := db.ReplicaDB.Table("homework_users hu").
		Select("hu.user_id, COUNT(DISTINCT hu.homework_id) as count").
		Joins("JOIN homeworks h ON h.id = hu.homework_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = hu.user_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.questions_completed < h.total_questions")
	inProgressQuery = inProgressQuery.Where("hrl.course_id = uc.course_id")

	if len(lessonIDs) > 0 {
		inProgressQuery = inProgressQuery.Where("l.id IN (?)", lessonIDs)
	}

	if req.ChapterID > 0 {
		inProgressQuery = inProgressQuery.Where("ch.id = ?", req.ChapterID)
	}

	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		inProgressQuery = inProgressQuery.Where("h.id IN (?)", homeworkIDs)
	}

	// Filter theo start_date và end_date
	if startDate != nil && endDate != nil {
		inProgressQuery = inProgressQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}

	if err := inProgressQuery.Group("hu.user_id").Find(&inProgressRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range inProgressRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.InProgressHomework = row.Count
		}
	}

	// 5) completed_homework (where homework_users.questions_completed >= homeworks.total_questions) join với lesson_schedules và weeks
	var completedRows []struct {
		UserID int64 `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	completedQuery := db.ReplicaDB.Table("homework_users hu").
		Select("hu.user_id, COUNT(DISTINCT hu.homework_id) as count").
		Joins("JOIN homeworks h ON h.id = hu.homework_id").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("JOIN lessons l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = hu.user_id").
		Where("h.deleted_at IS NULL AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.questions_completed >= h.total_questions")
	completedQuery = completedQuery.Where("hrl.course_id = uc.course_id")

	if len(lessonIDs) > 0 {
		completedQuery = completedQuery.Where("l.id IN (?)", lessonIDs)
	}

	if req.ChapterID > 0 {
		completedQuery = completedQuery.Where("ch.id = ?", req.ChapterID)
	}

	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		completedQuery = completedQuery.Where("h.id IN (?)", homeworkIDs)
	}

	// Filter theo end_date giống các thống kê khác (chỉ lấy homework được giao trước hoặc bằng end_date)
	if endDate != nil {
		completedQuery = completedQuery.Where("hrl.created_at <= ?", *endDate)
	}

	// Filter theo start_date và end_date
	if startDate != nil && endDate != nil {
		completedQuery = completedQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}

	if err := completedQuery.Group("hu.user_id").Find(&completedRows).Error; err != nil {
		return nil, 0, err
	}
	for _, row := range completedRows {
		if stats, ok := statsMap[row.UserID]; ok {
			stats.CompletedHomework = row.Count
		}
	}

	// 6) average_ratio (lấy trung bình hu.ratio) join với lesson_schedules và weeks
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
		Joins("JOIN lesson_schedules ls ON ls.course_id = c.id AND ls.lesson_id = l.id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Where("h.deleted_at IS NULL AND hrl.assigned_at IS NOT NULL AND uc.course_id = ?", req.CourseID).
		Where("hu.user_id IN (?)", userIDs).
		Where("hu.ratio IS NOT NULL")

	if len(lessonIDs) > 0 {
		avgRatioQuery = avgRatioQuery.Where("l.id IN (?)", lessonIDs)
	}

	if req.ChapterID > 0 {
		avgRatioQuery = avgRatioQuery.Where("ch.id = ?", req.ChapterID)
	}

	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		avgRatioQuery = avgRatioQuery.Where("h.id IN (?)", homeworkIDs)
	}

	// Filter theo start_date và end_date (đồng bộ với list-homeworks)
	if endDate != nil {
		avgRatioQuery = avgRatioQuery.Where("hrl.created_at <= ?", *endDate)
	}

	if startDate != nil && endDate != nil {
		avgRatioQuery = avgRatioQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
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
	// và tính completed_homework_ratio = completed_homework / total_assigned_homeworks
	for _, stats := range statsMap {
		stats.NotStartedHomework = stats.TotalAssignedHomeworks - stats.InProgressHomework - stats.CompletedHomework
		// Tính completed_homework_ratio
		if stats.TotalAssignedHomeworks > 0 {
			stats.CompletedHomeworkRatio = float64(stats.CompletedHomework) / float64(stats.TotalAssignedHomeworks)
		} else {
			stats.CompletedHomeworkRatio = 0
		}
	}

	// 8) Tính tổng số câu hỏi và câu đã hoàn thành
	// Chỉ tính khi homework_ids chỉ có 1 phần tử, nếu không thì set = 0
	if len(studentInfos) > 0 {
		// Kiểm tra nếu homework_ids chỉ có 1 phần tử
		if len(homeworkIDs) == 1 {
			type homeworkTotals struct {
				StudentID          int64
				TotalQuestions     int64
				QuestionsCompleted int64
			}

			var totals []homeworkTotals

			// Query đơn giản: lấy trực tiếp từ homeworks và homework_users
			// Khi chỉ có 1 homework_id, không cần filter phức tạp
			query := db.ReplicaDB.Table("homeworks h").
				Select(`
					uc.user_id as student_id,
					h.total_questions as total_questions,
					COALESCE(hu.questions_completed, 0) as questions_completed
				`).
				Joins("JOIN user_courses uc ON uc.course_id = ?", req.CourseID).
				Joins("LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = uc.user_id").
				Where("h.id = ?", homeworkIDs[0]).
				Where("h.deleted_at IS NULL").
				Where("uc.user_id IN (?)", userIDs)

			if err := query.Find(&totals).Error; err != nil {
				return nil, 0, err
			}

			for _, total := range totals {
				if stats, ok := statsMap[total.StudentID]; ok {
					stats.TotalQuestions = total.TotalQuestions
					stats.QuestionsCompleted = total.QuestionsCompleted
				}
			}
		} else {
			// Nếu không phải 1 homework_id, set total_questions và questions_completed = 0
			for _, stats := range statsMap {
				stats.TotalQuestions = 0
				stats.QuestionsCompleted = 0
			}
		}
	}

	// 9) Convert map to slice maintaining order
	students = make([]dto.DashboardTeacherHomeworkStudentStats, 0, len(studentInfos))
	for _, studentInfo := range studentInfos {
		if stats, ok := statsMap[studentInfo.UserID]; ok {
			students = append(students, *stats)
		}
	}

	// 10) Sắp xếp theo order_by nếu được truyền vào (sắp xếp giảm dần - từ cao xuống thấp)
	if req.OrderBy != "" {
		if req.OrderBy == "completed_homework" {
			// Sắp xếp giảm dần theo completed_homework_ratio (từ cao xuống thấp)
			sort.Slice(students, func(i, j int) bool {
				return students[i].CompletedHomeworkRatio > students[j].CompletedHomeworkRatio
			})
		} else if req.OrderBy == "average_score" {
			// Sắp xếp giảm dần theo average_ratio (từ cao xuống thấp)
			sort.Slice(students, func(i, j int) bool {
				return students[i].AverageRatio > students[j].AverageRatio
			})
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
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", req.HomeworkID).
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0")
	query = query.Where("hrl.course_id = user_courses.course_id")

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		query = query.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		query = query.Where("user_courses.course_id = ?", req.CourseID)
	}

	err := query.Distinct("user_courses.user_id").Count(&totalStudents).Error
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
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL AND homework_users.questions_completed < homeworks.total_questions", req.HomeworkID).
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0")
	inProgressQuery = inProgressQuery.Where("hrl.course_id = user_courses.course_id")

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		inProgressQuery = inProgressQuery.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		inProgressQuery = inProgressQuery.Where("user_courses.course_id = ?", req.CourseID)
	}

	err = inProgressQuery.Distinct("homework_users.user_id").Count(&inProgressCount).Error
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
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL AND homework_users.questions_completed >= homeworks.total_questions", req.HomeworkID).
		Where("hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0")
	completedQuery = completedQuery.Where("hrl.course_id = user_courses.course_id")

	// Add lesson_id filter if provided
	if req.LessonID > 0 {
		completedQuery = completedQuery.Where("lessons.id = ?", req.LessonID)
	}

	// Add course_id filter if provided
	if req.CourseID > 0 {
		completedQuery = completedQuery.Where("user_courses.course_id = ?", req.CourseID)
	}

	err = completedQuery.Distinct("homework_users.user_id").Count(&completedCount).Error
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

