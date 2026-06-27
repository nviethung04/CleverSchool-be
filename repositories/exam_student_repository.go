package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"time"

	"gorm.io/gorm"
)

type ExamStudentRepository interface {
	GetExamStudentsByExamID(examID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error)
	GetExamInfoByExamID(examID int64) (dto.ExamStudentInfo, error)
	GetExamByStudentRepo(userID, weekID int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error)
	GetExamsByLesson(lessonID, userID, courseID int64, exams *[]dto.GetExamByStudentExamQuery) error
	IsExamSubmitted(examID, userID int64) (bool, error)
	CountUnscoredManual(examID, userID int64) (int, error)
	GetHomeworksByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentHomeworkDTO, error)
	GetExercisesByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentExerciseDTO, error)
	IsExerciseSubmitted(exerciseID, userID int64) (bool, error)
	CountUnscoredManualExercise(exerciseID, userID int64) (int, error)
	CountHomeworkQuestionsCompleted(homeworkID, userID int64) (int32, error)
	GetExamComment(examID, studentID int64) (string, error)
	GetExamByStudentRepoWithDate(userID, weekID, courseID int64, startDate, endDate *int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error)
	GetAssessmentsByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentAssessmentDTO, error)
}

type examStudentRepository struct{}

func NewExamStudentRepository() ExamStudentRepository {
	return &examStudentRepository{}
}

func (r *examStudentRepository) GetExamStudentsByExamID(examID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error) {
	var total int64

	type examStudentRow struct {
		UserID        int64            `gorm:"column:user_id"`
		Name          string           `gorm:"column:name"`
		Username      string           `gorm:"column:username"`
		AvatarInfo    models.MediaInfo `gorm:"column:avatar_info"`
		IsSubmitted   bool             `gorm:"column:is_submitted"`
		SubmittedAt   int64            `gorm:"column:submitted_at"`
		Duration      int64            `gorm:"column:duration"`
		UnscoredCount int32            `gorm:"column:unscored_count"`
		Score         float64          `gorm:"column:score"`
		Ratio         float64          `gorm:"column:ratio"`
		CourseID      int64            `gorm:"column:course_id"`
	}

	buildQuery := func() *gorm.DB {
		q := db.ReplicaDB.Table("users").
			Joins("JOIN user_courses ON user_courses.user_id = users.id").
			Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
			Joins("JOIN chapters ON chapters.program_id = courses.program_id AND chapters.deleted_at IS NULL").
			Joins("JOIN lessons ON lessons.chapter_id = chapters.id AND lessons.deleted_at IS NULL").
			Joins("JOIN exam_ref_lessons erl ON erl.lesson_id = lessons.id").
			Joins("JOIN exams ON exams.id = erl.exam_id AND exams.deleted_at IS NULL").
			Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("exams.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", examID)
		if courseID > 0 {
			q = q.Where("erl.course_id = ?", courseID).
				Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
				Where("user_courses.course_id = ?", courseID)
		}
		return q
	}

	countQuery := buildQuery()
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []examStudentRow
	query := buildQuery().
		Select(`users.id as user_id, users.name, users.username, users.avatar_info,
			CASE WHEN exam_users.id IS NULL THEN false ELSE true END as is_submitted,
			COALESCE(CAST(extract(epoch from exam_users.created_at) AS BIGINT), 0) as submitted_at,
			COALESCE(exam_users."time", 0) as duration,
			COALESCE(unscored.count, 0) as unscored_count,
			COALESCE(exam_users.score, 0) as score,
			COALESCE(exam_users.ratio, 0) as ratio,
			erl.course_id as course_id`).
		Joins("LEFT JOIN exam_users ON exam_users.exam_id = exams.id AND exam_users.user_id = users.id").
		Joins(`LEFT JOIN (
			SELECT exam_id, user_id, COUNT(*) as count
			FROM exam_question_user_manual_scoring
			WHERE is_scored = false
			GROUP BY exam_id, user_id
		) as unscored ON unscored.exam_id = exams.id AND unscored.user_id = users.id`)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.ExamStudentItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, dto.ExamStudentItem{
			UserID:        row.UserID,
			Name:          row.Name,
			Username:      row.Username,
			Avatar:        row.AvatarInfo.Path,
			IsSubmitted:   row.IsSubmitted,
			SubmittedAt:   row.SubmittedAt,
			Duration:      row.Duration,
			UnscoredCount: row.UnscoredCount,
			Score:         row.Score,
			Ratio:         row.Ratio,
			CourseID:      row.CourseID,
		})
	}
	return result, total, nil
}

func (r *examStudentRepository) GetExamInfoByExamID(examID int64) (dto.ExamStudentInfo, error) {
	var info dto.ExamStudentInfo

	// Lấy thông tin exam, join với exam_ref_lessons để lấy assigned status
	err := db.ReplicaDB.Table("exams AS e").
		Select(`e.name, e.description,
			(CASE WHEN erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND erl.course_id > 0 THEN true ELSE false END) as is_assigned,
			e.status, e.max_score, e.time_limit,
			CAST(extract(epoch from e.created_at) AS BIGINT) as created_at,
			CAST(extract(epoch from e.deadline) AS BIGINT) as deadline`).
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Where("e.id = ? AND e.deleted_at IS NULL", examID).
		Limit(1).
		Scan(&info).Error
	if err != nil {
		return info, err
	}
	// Đếm tổng số học sinh (role_id = 3) liên quan đến exam qua course -> chapter -> lesson -> exam_ref_lessons
	var totalStudents int64
	err = db.ReplicaDB.Table("users").
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.program_id = courses.program_id AND chapters.deleted_at IS NULL").
		Joins("JOIN lessons ON lessons.chapter_id = chapters.id AND lessons.deleted_at IS NULL").
		Joins("JOIN exam_ref_lessons erl ON erl.lesson_id = lessons.id").
		Joins("JOIN exams ON exams.id = erl.exam_id AND exams.deleted_at IS NULL").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("exams.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", examID).
		Count(&totalStudents).Error
	if err != nil {
		return info, err
	}
	info.TotalStudents = int32(totalStudents)
	return info, nil
}

func (r *examStudentRepository) GetExamByStudentRepo(userID, weekID int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error) {
	// 1. Lấy danh sách course của học sinh
	var courses []struct {
		ID          int64
		Name        string
		Description string
		Status      bool
		SubjectName string
		Type        string
		Image       string
		Level       string
		Target      string
	}
	courseQuery := db.ReplicaDB.Table("user_courses").
		Select(`courses.id, courses.name, courses.description, courses.status, subjects.name as subject_name, courses.type, courses.image_info, courses.level, courses.target`).
		Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
		Joins("LEFT JOIN subjects ON subjects.id = courses.subject_id AND subjects.deleted_at IS NULL").
		Where("user_courses.user_id = ?", userID)
	if limit > 0 {
		courseQuery = courseQuery.Limit(limit)
	}
	if offset > 0 {
		courseQuery = courseQuery.Offset(offset)
	}
	if err := courseQuery.Scan(&courses).Error; err != nil {
		return nil, 0, err
	}
	var total int64
	db.ReplicaDB.Table("user_courses").Where("user_id = ?", userID).Count(&total)

	// 2. Lấy lessons theo từng course (ưu tiên lịch tuần; fallback chương trình / bài đã giao)
	var result []dto.GetExamByStudentCourseDTO
	weekIDs := []int64{}
	if weekID > 0 {
		weekIDs = append(weekIDs, weekID)
	}
	for _, c := range courses {
		lessons, err := r.getAssignedLessonsForCourse(c.ID, weekIDs)
		if err != nil {
			return nil, 0, err
		}
		var lessonDTOs []dto.GetExamByStudentLessonDTO
		for _, l := range lessons {
			lessonDTOs = append(lessonDTOs, dto.GetExamByStudentLessonDTO{
				ID:          l.ID,
				Title:       l.Title,
				Description: l.Description,
				Status:      l.Status,
			})
		}
		result = append(result, dto.GetExamByStudentCourseDTO{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Status:      c.Status,
			SubjectName: c.SubjectName,
			Type:        c.Type,
			Image:       c.Image,
			Level:       c.Level,
			Target:      c.Target,
			Lessons:     lessonDTOs,
		})
	}
	return result, total, nil
}

func (r *examStudentRepository) GetExamByStudentRepoWithDate(userID, weekID, courseID int64, startDate, endDate *int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error) {
	// Lấy danh sách tuần
	var weekIDs []int64
	if startDate != nil && endDate != nil {
		var weeks []models.Week
		weekQuery := db.ReplicaDB.Model(&models.Week{})
		weekQuery = weekQuery.Where("start_date >= ?", time.Unix(*startDate, 0))
		weekQuery = weekQuery.Where("end_date <= ?", time.Unix(*endDate, 0))
		if err := weekQuery.Find(&weeks).Error; err != nil {
			return nil, 0, err
		}
		for _, w := range weeks {
			weekIDs = append(weekIDs, w.ID)
		}
	} else if weekID > 0 {
		weekIDs = append(weekIDs, weekID)
	}

	// Lấy course của user
	var courses []struct {
		ID          int64
		Name        string
		Description string
		Status      bool
		SubjectName string
		Type        string
		Image       string
		Level       string
		Target      string
	}

	courseQuery := db.ReplicaDB.Table("user_courses").
		Select(`courses.id, courses.name, courses.description, courses.status, subjects.name as subject_name, courses.type, courses.image_info, courses.level, courses.target`).
		Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
		Joins("LEFT JOIN subjects ON subjects.id = courses.subject_id AND subjects.deleted_at IS NULL").
		Where("user_courses.user_id = ?", userID)

	if courseID > 0 {
		courseQuery = courseQuery.Where("courses.id = ?", courseID)
	}

	if limit > 0 {
		courseQuery = courseQuery.Limit(limit)
	}
	if offset > 0 {
		courseQuery = courseQuery.Offset(offset)
	}

	if err := courseQuery.Scan(&courses).Error; err != nil {
		return nil, 0, err
	}

	// Đếm total khóa học user đang học (phục vụ phân trang)
	var totalQuery = db.ReplicaDB.Table("user_courses").Where("user_id = ?", userID)
	if courseID > 0 {
		totalQuery = totalQuery.Where("course_id = ?", courseID)
	}
	var total int64
	totalQuery.Count(&total)

	// Lấy lessons theo từng course (ưu tiên lịch tuần; fallback bài đã giao qua *_ref_lessons)
	var result []dto.GetExamByStudentCourseDTO
	for _, c := range courses {
		lessons, err := r.getAssignedLessonsForCourse(c.ID, weekIDs)
		if err != nil {
			return nil, 0, err
		}

		var lessonDTOs []dto.GetExamByStudentLessonDTO
		for _, l := range lessons {
			lessonDTOs = append(lessonDTOs, dto.GetExamByStudentLessonDTO{
				ID:          l.ID,
				Title:       l.Title,
				Description: l.Description,
				Status:      l.Status,
			})
		}

		result = append(result, dto.GetExamByStudentCourseDTO{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Status:      c.Status,
			SubjectName: c.SubjectName,
			Type:        c.Type,
			Image:       c.Image,
			Level:       c.Level,
			Target:      c.Target,
			Lessons:     lessonDTOs,
		})
	}

	return result, total, nil
}

type examStudentLessonRow struct {
	ID          int64
	Title       string
	Description string
	Status      bool
}

func (r *examStudentRepository) getAssignedLessonsForCourse(courseID int64, weekIDs []int64) ([]examStudentLessonRow, error) {
	var lessons []examStudentLessonRow

	if len(weekIDs) > 0 {
		q := db.ReplicaDB.Table("lessons").
			Select("DISTINCT lessons.id, lessons.title, lessons.description, lessons.status").
			Joins("JOIN lesson_schedules ON lesson_schedules.lesson_id = lessons.id").
			Where("lesson_schedules.course_id = ? AND lessons.deleted_at IS NULL", courseID).
			Where("lesson_schedules.week_id IN ?", weekIDs)
		if err := q.Scan(&lessons).Error; err != nil {
			return nil, err
		}
		if len(lessons) > 0 {
			return lessons, nil
		}
	}

	// Tất cả bài học thuộc chương trình khóa (program_id hoặc chapter.course_id legacy)
	err := db.ReplicaDB.Table("lessons l").
		Select("DISTINCT l.id, l.title, l.description, l.status").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.deleted_at IS NULL AND (c.program_id = ch.program_id OR ch.course_id = c.id)").
		Where("c.id = ? AND l.deleted_at IS NULL", courseID).
		Order("l.id ASC").
		Scan(&lessons).Error
	if err != nil {
		return nil, err
	}
	if len(lessons) > 0 {
		return lessons, nil
	}

	// Fallback: bài học có bài đã giao qua *_ref_lessons (khi không map được chương trình)
	err = db.ReplicaDB.Table("lessons l").
		Select("DISTINCT l.id, l.title, l.description, l.status").
		Where("l.deleted_at IS NULL").
		Where(`l.id IN (
			SELECT hrl.lesson_id FROM homework_ref_lessons hrl
			WHERE hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0
			  AND (hrl.course_id IS NULL OR hrl.course_id = 0 OR hrl.course_id = ?)
			UNION
			SELECT erl.lesson_id FROM exam_ref_lessons erl
			WHERE erl.assigned_by IS NOT NULL AND erl.assigned_by > 0
			  AND (erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = ?)
			UNION
			SELECT xrl.lesson_id FROM exercise_ref_lessons xrl
			WHERE xrl.assigned_by IS NOT NULL AND xrl.assigned_by > 0
			  AND (xrl.course_id IS NULL OR xrl.course_id = 0 OR xrl.course_id = ?)
		)`, courseID, courseID, courseID).
		Order("l.id ASC").
		Scan(&lessons).Error
	return lessons, err
}

func (r *examStudentRepository) GetExamsByLesson(lessonID, userID, courseID int64, exams *[]dto.GetExamByStudentExamQuery) error {
	return db.ReplicaDB.Table("exams AS e").
		Select("e.id, e.name, e.status, e.description, e.cover_image_info, CAST(extract(epoch from e.deadline) AS BIGINT) as deadline").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Where("erl.lesson_id = ? AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND e.deleted_at IS NULL", lessonID).
		Where("(erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = ?)", courseID).
		Order("e.id ASC").
		Scan(exams).Error
}

func (r *examStudentRepository) IsExamSubmitted(examID, userID int64) (bool, error) {
	var count int64
	err := db.ReplicaDB.Table("exam_users").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *examStudentRepository) CountUnscoredManual(examID, userID int64) (int, error) {
	var count int64
	err := db.ReplicaDB.Table("exam_question_user_manual_scoring").
		Where("exam_id = ? AND user_id = ? AND is_scored = false", examID, userID).
		Count(&count).Error
	return int(count), err
}

func (r *examStudentRepository) GetHomeworksByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentHomeworkDTO, error) {
	var homeworks []struct {
		ID                      int64
		Name                    string
		Status                  int32
		Description             string
		CoverImage              string
		TotalQuestion           int32
		QuestionCompleted       int32
		LastQuestionIDCompleted int64
	}
	err := db.ReplicaDB.Table("homeworks h").
		Select(`h.id, h.name, h.status, h.description, h.cover_image_info,
			COALESCE(hq.total_question, 0) as total_question,
			COALESCE(hu.questions_completed, 0) as question_completed,
			COALESCE(hu.last_question_id_completed, 0) as last_question_id_completed`).
		Joins(`JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id AND hrl.lesson_id = ? AND hrl.assigned_by IS NOT NULL AND hrl.assigned_by > 0 AND (hrl.course_id IS NULL OR hrl.course_id = 0 OR hrl.course_id = ?)`, lessonID, courseID).
		Joins(`LEFT JOIN (
			SELECT assignment_id as homework_id, jsonb_array_length(questions) as total_question
			FROM cloned_questions
			WHERE assignment_type = 'homework'
		) as hq ON hq.homework_id = h.id`).
		Joins(`LEFT JOIN homework_users hu ON hu.homework_id = h.id AND hu.user_id = ?`, userID).
		Where("h.deleted_at IS NULL").
		Order("h.id ASC").
		Scan(&homeworks).Error
	if err != nil {
		return nil, err
	}
	var result []dto.GetExamByStudentHomeworkDTO
	for _, h := range homeworks {
		result = append(result, dto.GetExamByStudentHomeworkDTO{
			ID:                      h.ID,
			Name:                    h.Name,
			Status:                  h.Status,
			Description:             h.Description,
			TotalQuestion:           h.TotalQuestion,
			QuestionCompleted:       h.QuestionCompleted,
			CoverImage:              h.CoverImage,
			LastQuestionIDCompleted: h.LastQuestionIDCompleted,
		})
	}
	return result, nil
}

func (r *examStudentRepository) GetExercisesByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentExerciseDTO, error) {
	var rows []struct {
		ID          int64
		Name        string
		Status      int32
		Description string
		CoverImage  string
		Deadline    int64
	}
	err := db.ReplicaDB.Table("exercises e").
		Select("e.id, e.name, e.status, e.description, e.cover_image_info, CAST(extract(epoch from e.deadline) AS BIGINT) as deadline").
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id AND erl.lesson_id = ? AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND (erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = ?)", lessonID, courseID).
		Where("e.deleted_at IS NULL").
		Order("e.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	var result []dto.GetExamByStudentExerciseDTO
	for _, r0 := range rows {
		result = append(result, dto.GetExamByStudentExerciseDTO{
			ID:          r0.ID,
			Name:        r0.Name,
			Status:      r0.Status,
			Description: r0.Description,
			CoverImage:  r0.CoverImage,
			Deadline:    r0.Deadline,
		})
	}
	return result, nil
}

func (r *examStudentRepository) GetAssessmentsByLesson(lessonID, userID, courseID int64) ([]dto.GetExamByStudentAssessmentDTO, error) {
	var rows []struct {
		ID          int64
		Name        string
		Description string
		Type        string
	}
	err := db.ReplicaDB.Table("assessments a").
		Select("DISTINCT a.id, a.name, a.description, a.type").
		Joins("JOIN assessment_ref_lessons arl ON arl.assessment_id = a.id").
		Where("arl.lesson_id = ? AND arl.assigned_by IS NOT NULL AND arl.assigned_by > 0 AND a.deleted_at IS NULL", lessonID).
		Where("(arl.course_id IS NULL OR arl.course_id = 0 OR arl.course_id = ?)", courseID).
		Order("a.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.GetExamByStudentAssessmentDTO, 0, len(rows))
	for _, row := range rows {
		var scoreCount int64
		var isScored bool
		_ = db.ReplicaDB.Table("assessment_scores").
			Where("assessment_id = ? AND user_id = ? AND course_id = ?", row.ID, userID, courseID).
			Count(&scoreCount).Error
		isSubmitted := scoreCount > 0
		if isSubmitted {
			_ = db.ReplicaDB.Table("assessment_scores").
				Select("is_scored").
				Where("assessment_id = ? AND user_id = ? AND course_id = ?", row.ID, userID, courseID).
				Limit(1).
				Scan(&isScored).Error
		}
		var isPublished bool
		_ = db.ReplicaDB.Table("assessment_publishes").
			Select("publish").
			Where("assessment_id = ? AND course_id = ?", row.ID, courseID).
			Limit(1).
			Scan(&isPublished).Error

		result = append(result, dto.GetExamByStudentAssessmentDTO{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Type:        row.Type,
			IsSubmitted: isSubmitted,
			IsScored:    isScored,
			IsPublished: isPublished,
		})
	}
	return result, nil
}

func (r *examStudentRepository) IsExerciseSubmitted(exerciseID, userID int64) (bool, error) {
	var count int64
	err := db.ReplicaDB.Table("exercise_users").Where("exercise_id = ? AND user_id = ?", exerciseID, userID).Count(&count).Error
	return count > 0, err
}

func (r *examStudentRepository) CountUnscoredManualExercise(exerciseID, userID int64) (int, error) {
	var count int64
	err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").
		Where("exercise_id = ? AND user_id = ? AND is_scored = false", exerciseID, userID).
		Count(&count).Error
	return int(count), err
}

func (r *examStudentRepository) CountHomeworkQuestionsCompleted(homeworkID, userID int64) (int32, error) {
	var count int64
	err := db.ReplicaDB.Table("homework_question_users").Where("homework_id = ? AND user_id = ?", homeworkID, userID).Count(&count).Error
	return int32(count), err
}

func (r *examStudentRepository) GetExamComment(examID, studentID int64) (string, error) {
	var content string
	err := db.ReplicaDB.Table("exam_comments").
		Select("content").
		Where("exam_id = ? AND student_id = ?", examID, studentID).
		Order("created_at desc").
		Limit(1).
		Scan(&content).Error
	return content, err
}
