package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"time"

	"github.com/gin-gonic/gin"
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
	GetExamByStudentRepoWithDate(c *gin.Context, userID, weekID, courseID int64, startDate, endDate *int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error)
}

type examStudentRepository struct{}

func NewExamStudentRepository() ExamStudentRepository {
	return &examStudentRepository{}
}

func (r *examStudentRepository) GetExamStudentsByExamID(examID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error) {
	var result []dto.ExamStudentItem
	var total int64

	countQuery := db.ReplicaDB.Table("exams").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id AND lessons.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id AND chapters.deleted_at IS NULL").
		Joins("JOIN courses ON courses.program_id = chapters.program_id AND courses.deleted_at IS NULL").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN users ON users.id = user_courses.user_id AND urr.role_id = 3 AND users.deleted_at IS NULL").
		Where("exams.id = ? AND exams.deleted_at IS NULL", examID)

	if courseID > 0 {
		countQuery = countQuery.Where("erl.course_id = ?", courseID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := db.ReplicaDB.Table("exams").
		Select(`users.id as user_id, users.name, users.username, users.avatar_info,
			CASE WHEN exam_users.id IS NULL THEN false ELSE true END as is_submitted,
			CAST(extract(epoch from exam_users.created_at) AS BIGINT) as submitted_at,
			exam_users.time as duration,
			COALESCE(unscored.count, 0) as unscored_count,
			exam_users.score as score,
			exam_users.ratio as ratio,
			erl.course_id as course_id`).
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id AND lessons.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id AND chapters.deleted_at IS NULL").
		Joins("JOIN courses ON courses.program_id = chapters.program_id AND courses.deleted_at IS NULL").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN users ON users.id = user_courses.user_id AND urr.role_id = 3 AND users.deleted_at IS NULL").
		Joins("LEFT JOIN exam_users ON exam_users.exam_id = exams.id AND exam_users.user_id = users.id").
		Joins(`LEFT JOIN (
			SELECT exam_id, user_id, COUNT(*) as count
			FROM exam_question_user_manual_scoring
			WHERE is_scored = false
			GROUP BY exam_id, user_id
		) as unscored ON unscored.exam_id = exams.id AND unscored.user_id = users.id`).
		Where("exams.id = ? AND exams.deleted_at IS NULL", examID)

	if courseID > 0 {
		query = query.Where("erl.course_id = ?", courseID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, 0, err
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

	// 2. Lấy lessons theo từng course trong tuần
	var result []dto.GetExamByStudentCourseDTO
	for _, c := range courses {
		var lessons []struct {
			ID          int64
			Title       string
			Description string
			Status      bool
		}
		lessonQuery := db.ReplicaDB.Table("lessons").
			Select("lessons.id, lessons.title, lessons.description, lessons.status").
			Joins("JOIN lesson_schedules ON lesson_schedules.lesson_id = lessons.id").
			Where("lesson_schedules.course_id = ? AND lesson_schedules.week_id = ? AND lessons.deleted_at IS NULL", c.ID, weekID)
		if err := lessonQuery.Scan(&lessons).Error; err != nil {
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

func (r *examStudentRepository) GetExamByStudentRepoWithDate(c *gin.Context, userID, weekID, courseID int64, startDate, endDate *int64, limit, offset int) ([]dto.GetExamByStudentCourseDTO, int64, error) {
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

	// memberType, _ := c.Get("memberType")

	// if len(courses) == 0 && memberType.(string) == models.MemberTypeExternal {
	// 	courseIds := config.LoadConfig().PublicCourseIds

	// 	if courseID > 0 {
	// 		courseIds = []int64{courseID}
	// 	}

	// 	db.ReplicaDB.Table("courses").
	// 		Where("courses.id IN ?", courseIds).
	// 		Joins("LEFT JOIN subjects ON subjects.id = courses.subject_id AND subjects.deleted_at IS NULL").
	// 		Select(`courses.id, courses.name, courses.description, courses.status, subjects.name as subject_name, courses.type, courses.image_info, courses.level, courses.target`).
	// 		Scan(&courses)

	// 	total = int64(len(courseIds))
	// }

	// Lấy lessons theo từng course trong tuần
	var result []dto.GetExamByStudentCourseDTO
	for _, c := range courses {
		var lessons []struct {
			ID          int64
			Title       string
			Description string
			Status      bool
		}
		lessonQuery := db.ReplicaDB.Table("lessons").
			Select("lessons.id, lessons.title, lessons.description, lessons.status").
			Joins("JOIN lesson_schedules ON lesson_schedules.lesson_id = lessons.id").
			Where("lesson_schedules.course_id = ? AND lessons.deleted_at IS NULL", c.ID)

		if len(weekIDs) > 0 {
			lessonQuery = lessonQuery.Where("lesson_schedules.week_id IN ?", weekIDs)
		}

		if err := lessonQuery.Scan(&lessons).Error; err != nil {
			return nil, 0, err
		}

		var lessonDTOs []dto.GetExamByStudentLessonDTO
		for _, l := range lessons {
			// Lấy exams theo lesson qua bảng trung gian exam_ref_lessons
			var exams []dto.GetExamByStudentExamDTO
			err := db.ReplicaDB.Table("exams AS e").
				Select("e.id, e.name, e.status::int as status, e.description, e.cover_image_info, CAST(extract(epoch from e.deadline) AS BIGINT) as deadline, e.type").
				Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
				Where("erl.lesson_id = ? AND e.deleted_at IS NULL", l.ID).
				Where("erl.course_id = ?", courseID).
				Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND e.deleted_at IS NULL").
				Order("e.id ASC").
				Scan(&exams).Error

			if err != nil {
				return nil, 0, err
			}

			lessonDTOs = append(lessonDTOs, dto.GetExamByStudentLessonDTO{
				ID:          l.ID,
				Title:       l.Title,
				Description: l.Description,
				Status:      l.Status,
				Exams:       exams,
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

func (r *examStudentRepository) GetExamsByLesson(lessonID, userID, courseID int64, exams *[]dto.GetExamByStudentExamQuery) error {
	return db.ReplicaDB.Table("exams AS e").
		Select("e.id, e.name, e.status, e.description, e.cover_image_info, CAST(extract(epoch from e.deadline) AS BIGINT) as deadline, e.type").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Where("erl.lesson_id = ? AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND e.deleted_at IS NULL AND erl.course_id = ?", lessonID, courseID).
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
		QuestionForm            string
		IsSubmitted             bool
	}
	err := db.ReplicaDB.Table("homeworks h").
		Select(`h.id, h.name, h.status, h.description, h.cover_image_info, h.question_form,
			COALESCE(hq.total_question, 0) AS total_question,
			COALESCE(hu.questions_completed, 0) AS question_completed,
			COALESCE(hu.last_question_id_completed, 0) AS last_question_id_completed,
			CASE WHEN hu.id IS NOT NULL THEN TRUE ELSE FALSE END AS is_submitted`).
		Joins(`JOIN homework_ref_lessons hrl
			ON hrl.homework_id = h.id
			AND hrl.lesson_id = ?
			AND hrl.assigned_by IS NOT NULL
			AND hrl.assigned_by > 0
			AND hrl.course_id = ?`, lessonID, courseID).
		Joins(`LEFT JOIN (
			SELECT assignment_id AS homework_id, jsonb_array_length(questions) AS total_question
			FROM cloned_questions
			WHERE assignment_type = 'homework'
		) AS hq ON hq.homework_id = h.id`).
		Joins(`LEFT JOIN homework_users hu
			ON hu.homework_id = h.id
			AND hu.user_id = ?`, userID).
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
			QuestionForm:            h.QuestionForm,
			IsSubmitted:             h.IsSubmitted,
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
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id AND erl.lesson_id = ? AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 AND erl.course_id = ?", lessonID, courseID).
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
		Where("exams_id = ? AND student_id = ?", examID, studentID).
		Order("created_at desc").
		Limit(1).
		Scan(&content).Error
	return content, err
}

