package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
)

type ExerciseStudentRepository interface {
	GetExerciseStudentsByExerciseID(exerciseID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error)
	GetExerciseInfoByExerciseID(exerciseID int64) (dto.ExamStudentInfo, error)
	GetExerciseComment(exerciseID, userID int64) (string, error)
}

type exerciseStudentRepository struct{}

func NewExerciseStudentRepository() ExerciseStudentRepository { return &exerciseStudentRepository{} }

func (r *exerciseStudentRepository) GetExerciseStudentsByExerciseID(exerciseID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error) {
	var result []dto.ExamStudentItem
	var total int64

	countQuery := db.ReplicaDB.Table("exercises").
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id AND lessons.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id AND chapters.deleted_at IS NULL").
		Joins("JOIN courses ON courses.program_id = chapters.program_id AND courses.deleted_at IS NULL").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN users ON users.id = user_courses.user_id AND urr.role_id = 3 AND users.deleted_at IS NULL").
		Where("exercises.id = ? AND exercises.deleted_at IS NULL", exerciseID)

	if courseID > 0 {
		countQuery = countQuery.Where("erl.course_id = ?", courseID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := db.ReplicaDB.Table("exercises").
		Select(`users.id as user_id, users.name, users.username, users.avatar_info,
            CASE WHEN exercise_users.id IS NULL THEN false ELSE true END as is_submitted,
            CAST(extract(epoch from exercise_users.created_at) AS BIGINT) as submitted_at,
            exercise_users.time as duration,
            COALESCE(unscored.count, 0) as unscored_count,
            exercise_users.score as score,
            exercise_users.ratio as ratio,
            erl.course_id as course_id`).
		Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id AND lessons.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id AND chapters.deleted_at IS NULL").
		Joins("JOIN courses ON courses.program_id = chapters.program_id AND courses.deleted_at IS NULL").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN users ON users.id = user_courses.user_id AND urr.role_id = 3 AND users.deleted_at IS NULL").
		Joins("LEFT JOIN exercise_users ON exercise_users.exercise_id = exercises.id AND exercise_users.user_id = users.id").
		Joins(`LEFT JOIN (
            SELECT exercise_id, user_id, COUNT(*) as count
            FROM exercise_question_user_manual_scoring
            WHERE is_scored = false
            GROUP BY exercise_id, user_id
        ) as unscored ON unscored.exercise_id = exercises.id AND unscored.user_id = users.id`).
		Where("exercises.id = ? AND exercises.deleted_at IS NULL", exerciseID)

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

func (r *exerciseStudentRepository) GetExerciseInfoByExerciseID(exerciseID int64) (dto.ExamStudentInfo, error) {
	var info dto.ExamStudentInfo
	// Thông tin exercise + trạng thái assigned
	err := db.ReplicaDB.Table("exercises AS e").
		Select(`e.name, e.description,
            (CASE WHEN erl.assigned_by IS NOT NULL AND erl.assigned_by > 0 THEN true ELSE false END) as is_assigned,
            e.status, e.max_score, e.time_limit,
            CAST(extract(epoch from e.created_at) AS BIGINT) as created_at,
            CAST(extract(epoch from e.deadline) AS BIGINT) as deadline`).
		Joins("LEFT JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id").
		Where("e.id = ? AND e.deleted_at IS NULL", exerciseID).
		Limit(1).
		Scan(&info).Error
	if err != nil {
		return info, err
	}

	// Tổng số học sinh liên quan
	var totalStudents int64
	err = db.ReplicaDB.Table("users").
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.program_id = courses.program_id AND chapters.deleted_at IS NULL").
		Joins("JOIN lessons ON lessons.chapter_id = chapters.id AND lessons.deleted_at IS NULL").
		Joins("JOIN exercise_ref_lessons erl ON erl.lesson_id = lessons.id").
		Joins("JOIN exercises ON exercises.id = erl.exercise_id AND exercises.deleted_at IS NULL").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("exercises.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", exerciseID).
		Count(&totalStudents).Error
	if err != nil {
		return info, err
	}
	info.TotalStudents = int32(totalStudents)
	return info, nil
}

func (r *exerciseStudentRepository) GetExerciseComment(exerciseID, userID int64) (string, error) {
	var content string
	err := db.ReplicaDB.Table("exercise_comments").
		Select("content").
		Where("exercises_id = ? AND student_id = ?", exerciseID, userID).
		Order("created_at desc").
		Limit(1).
		Scan(&content).Error
	return content, err
}

