package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"

	"gorm.io/gorm/clause"
)

type AssessmentRefLessonRepository interface {
	GetProgramIDByLessonID(lessonID int64) (int64, error)
	GetCoursesByProgramID(programID int64) ([]models.Course, error)
	DeleteByAssessmentAndLesson(assessmentID, lessonID int64) error
	DeleteByAssessmentLessonAndCourses(assessmentID, lessonID int64, courseIDs []int64) error
	AssignCoursesToAssessment(assessmentID, lessonID int64, courseIDs []int64, assignedBy int64) error
}

type assessmentRefLessonRepository struct{}

func NewAssessmentRefLessonRepository() AssessmentRefLessonRepository {
	return &assessmentRefLessonRepository{}
}

func (r *assessmentRefLessonRepository) GetProgramIDByLessonID(lessonID int64) (int64, error) {
	var programID int64
	err := db.ReplicaDB.
		Table("lessons l").
		Select("ch.program_id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Where("l.id = ? AND l.deleted_at IS NULL", lessonID).
		Scan(&programID).Error
	if err != nil {
		return 0, err
	}
	return programID, nil
}

func (r *assessmentRefLessonRepository) GetCoursesByProgramID(programID int64) ([]models.Course, error) {
	var courses []models.Course
	err := db.ReplicaDB.
		Where("program_id = ? AND deleted_at IS NULL", programID).
		Find(&courses).Error
	return courses, err
}

func (r *assessmentRefLessonRepository) DeleteByAssessmentAndLesson(assessmentID, lessonID int64) error {
	return db.MasterDB.
		Where("assessment_id = ? AND lesson_id = ?", assessmentID, lessonID).
		Delete(&models.AssessmentRefLesson{}).Error
}

func (r *assessmentRefLessonRepository) DeleteByAssessmentLessonAndCourses(assessmentID, lessonID int64, courseIDs []int64) error {
	if len(courseIDs) == 0 {
		return nil
	}
	return db.MasterDB.
		Where("assessment_id = ? AND lesson_id = ? AND course_id IN (?)", assessmentID, lessonID, courseIDs).
		Delete(&models.AssessmentRefLesson{}).Error
}

func (r *assessmentRefLessonRepository) AssignCoursesToAssessment(assessmentID, lessonID int64, courseIDs []int64, assignedBy int64) error {
	if len(courseIDs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	refs := make([]models.AssessmentRefLesson, 0, len(courseIDs))

	for _, courseID := range courseIDs {
		refs = append(refs, models.AssessmentRefLesson{
			AssessmentId: assessmentID,
			LessonId:     lessonID,
			CourseId:     courseID,
			AssignedAt:   &now,
			AssignedBy:   &assignedBy,
		})
	}

	// Sử dụng ON CONFLICT DO NOTHING để chỉ thêm những record chưa tồn tại
	// Nếu đã tồn tại (cùng bộ 3 assessment_id, lesson_id, course_id) → giữ nguyên
	err := db.MasterDB.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "assessment_id"},
				{Name: "lesson_id"},
				{Name: "course_id"},
			},
			DoNothing: true,
		}).
		Create(&refs).Error

	return err
}
