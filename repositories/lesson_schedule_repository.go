package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"strconv"
	"strings"
	"time"
)

type LessonScheduleRepository interface {
	GetSchedules(filters map[string]interface{}) ([]models.LessonSchedule, error)
	GetSchedulesByUser(userID int, filters map[string]interface{}) ([]models.LessonSchedule, error)
	UpdateCourseIdByChapter(chapterId, courseId, oldCourseId int64) error
	UpdateCourseIdByLesson(lessonId, chapterId, oldChapterId int64) error
}

type lessonScheduleRepository struct{}

func NewLessonScheduleRepository() LessonScheduleRepository {
	return &lessonScheduleRepository{}
}

func (r *lessonScheduleRepository) GetSchedulesByUser(userID int, filters map[string]interface{}) ([]models.LessonSchedule, error) {
	var schedules []models.LessonSchedule

	query := db.ReplicaDB.
		Joins("JOIN lessons ON lessons.id = lesson_schedules.lesson_id").
		Joins("JOIN courses ON courses.id = lesson_schedules.course_id").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Where("user_courses.user_id = ?", userID).
		Where("lessons.deleted_at IS NULL").
		Where("courses.deleted_at IS NULL").
		Where("chapters.program_id = courses.program_id")

	if start, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("lesson_schedules.scheduled_date >= ?", start)
	}
	if end, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("lesson_schedules.scheduled_date <= ?", end)
	}
	if weekID, ok := filters["week_id"].(int64); ok && weekID > 0 {
		query = query.Where("lesson_schedules.week_id = ?", weekID)
	}

	if weekDate, ok := filters["week_date"].(time.Time); ok {
		query = query.Joins("JOIN weeks ON weeks.id = lesson_schedules.week_id").
			Where("weeks.start_date <= ? AND weeks.end_date >= ?", weekDate, weekDate)
	}

	err := query.
		Preload("Week").
		Preload("Lesson").
		Preload("Course").
		Preload("Lesson.Chapter").
		Preload("Lesson.LessonPlans").
		Preload("Lesson.LessonPlans.Author").
		Find(&schedules).Error

	return schedules, err
}

func (r *lessonScheduleRepository) GetSchedules(filters map[string]interface{}) ([]models.LessonSchedule, error) {
	var schedules []models.LessonSchedule

	query := db.MasterDB.
		Joins("JOIN lessons ON lessons.id = lesson_schedules.lesson_id").
		Joins("JOIN courses ON courses.id = lesson_schedules.course_id").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id").
		Where("lessons.deleted_at IS NULL").
		Where("courses.deleted_at IS NULL").
		Where("chapters.program_id = courses.program_id")

	if start, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("lesson_schedules.scheduled_date >= ?", start)
	}
	if end, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("lesson_schedules.scheduled_date <= ?", end)
	}
	if weekID, ok := filters["week_id"].(int64); ok && weekID > 0 {
		query = query.Where("lesson_schedules.week_id = ?", weekID)
	}

	if start, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("lesson_schedules.scheduled_date >= ?", start)
	}

	if userID, ok := filters["user_id"].(int); ok && userID > 0 {
		query = query.Joins("JOIN user_courses ON user_courses.course_id = courses.id").
			Where("user_courses.user_id = ?", userID)
	}

	if courseID, ok := filters["course_id"].(int64); ok && courseID > 0 {
		query = query.
			Where("lesson_schedules.course_id = ?", courseID)
	}

	if idStr, ok := filters["lesson_ids"].(string); ok {
		if idStr != "" {
			if ids, err := parseIDList(idStr); err == nil && len(ids) > 0 {
				query = query.Where("lesson_schedules.lesson_id IN ?", ids)
			} else {
				query = query.Where("lesson_schedules.lesson_id = ?", 0)
			}
		} else {
			query = query.Where("lesson_schedules.lesson_id = ?", 0)
		}
	}

	err := query.
		Preload("Week").
		Preload("Lesson").
		Preload("Course").
		Preload("Lesson.Chapter").
		Preload("Lesson.LessonPlans").
		Preload("Lesson.LessonPlans.Author").
		Find(&schedules).Error

	return schedules, err
}

func parseIDList(idStr string) ([]int64, error) {
	parts := strings.Split(idStr, ",")
	var ids []int64

	for _, part := range parts {
		if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
			ids = append(ids, id)
		} else {
			return nil, err
		}
	}

	return ids, nil
}

func (r *lessonScheduleRepository) UpdateCourseIdByChapter(chapterId, courseId, oldCourseId int64) error {
	if chapterId == 0 || courseId == 0 || oldCourseId == 0 {
		return nil
	}

	// 1. Lấy danh sách lesson_id từ chapterId mới
	var lessonIDs []int64
	err := db.MasterDB.
		Model(&models.Lesson{}).
		Where("chapter_id = ?", chapterId).
		Pluck("id", &lessonIDs).Error
	if err != nil {
		return err
	}

	if len(lessonIDs) == 0 {
		return nil // Không có lesson nào => không cần update
	}

	// 2. Cập nhật course_id trong lesson_schedules
	err = db.MasterDB.
		Model(&models.LessonSchedule{}).
		Where("lesson_id IN ?", lessonIDs).
		Where("course_id = ?", oldCourseId).
		Update("course_id", courseId).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *lessonScheduleRepository) UpdateCourseIdByLesson(lessonId, chapterId, oldChapterId int64) error {
	if lessonId == 0 || chapterId == 0 || oldChapterId == 0 {
		return nil
	}

	// 1. Lấy course_id mới từ chapterId
	var newCourseID int64
	err := db.MasterDB.
		Model(&models.Chapter{}).
		Select("course_id").
		Where("id = ?", chapterId).
		Scan(&newCourseID).Error
	if err != nil {
		return err
	}

	// 2. Lấy course_id cũ từ oldChapterId
	var oldCourseID int64
	err = db.MasterDB.
		Model(&models.Chapter{}).
		Select("course_id").
		Where("id = ?", oldChapterId).
		Scan(&oldCourseID).Error
	if err != nil {
		return err
	}

	// 3. Update course_id nếu lesson_schedule đang dùng course_id cũ
	err = db.MasterDB.
		Model(&models.LessonSchedule{}).
		Where("lesson_id = ?", lessonId).
		Where("course_id = ?", oldCourseID).
		Update("course_id", newCourseID).Error
	if err != nil {
		return err
	}

	return nil
}
