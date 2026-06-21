package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"

	"gorm.io/gorm"
)

type ExerciseStudentRepository interface {
	GetExerciseStudentsByExerciseID(exerciseID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error)
	GetExerciseInfoByExerciseID(exerciseID int64) (dto.ExamStudentInfo, error)
	GetExerciseComment(exerciseID, userID int64) (string, error)
}

type exerciseStudentRepository struct{}

func NewExerciseStudentRepository() ExerciseStudentRepository { return &exerciseStudentRepository{} }

func (r *exerciseStudentRepository) GetExerciseStudentsByExerciseID(exerciseID int64, courseID int64, limit, offset int) ([]dto.ExamStudentItem, int64, error) {
	var total int64

	type exerciseStudentRow struct {
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
			Joins("JOIN exercise_ref_lessons erl ON erl.lesson_id = lessons.id").
			Joins("JOIN exercises ON exercises.id = erl.exercise_id AND exercises.deleted_at IS NULL").
			Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("exercises.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", exerciseID)
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

	var rows []exerciseStudentRow
	query := buildQuery().
		Select(`users.id as user_id, users.name, users.username, users.avatar_info,
            CASE WHEN exercise_users.id IS NULL THEN false ELSE true END as is_submitted,
            COALESCE(CAST(extract(epoch from exercise_users.created_at) AS BIGINT), 0) as submitted_at,
            COALESCE(exercise_users."time", 0) as duration,
            COALESCE(unscored.count, 0) as unscored_count,
            COALESCE(exercise_users.score, 0) as score,
            COALESCE(exercise_users.ratio, 0) as ratio,
            erl.course_id as course_id`).
		Joins("LEFT JOIN exercise_users ON exercise_users.exercise_id = exercises.id AND exercise_users.user_id = users.id").
		Joins(`LEFT JOIN (
            SELECT exercise_id, user_id, COUNT(*) as count
            FROM exercise_question_user_manual_scoring
            WHERE is_scored = false
            GROUP BY exercise_id, user_id
        ) as unscored ON unscored.exercise_id = exercises.id AND unscored.user_id = users.id`)

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
		Where("exercise_id = ? AND student_id = ?", exerciseID, userID).
		Order("created_at desc").
		Limit(1).
		Scan(&content).Error
	return content, err
}
