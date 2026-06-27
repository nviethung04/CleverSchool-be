package repositories

import (
	"be-lms/database/db"
	"be-lms/models"

	"gorm.io/gorm"
)

type DashboardStudentAssessmentRepository interface {
	ListAssessmentsForStudent(userID int64, courseID int64, subjectID *int64, assessmentType string, assessmentID *int64, limit, offset int) ([]models.Assessment, int64, error)
}

type dashboardStudentAssessmentRepository struct{}

func NewDashboardStudentAssessmentRepository() DashboardStudentAssessmentRepository {
	return &dashboardStudentAssessmentRepository{}
}

func (r *dashboardStudentAssessmentRepository) baseQuery(userID int64, courseID int64) *gorm.DB {
	query := db.ReplicaDB.Model(&models.Assessment{}).
		Select("DISTINCT assessments.*").
		Joins("JOIN assessment_ref_lessons arl ON arl.assessment_id = assessments.id").
		Joins("JOIN lessons l ON l.id = arl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL").
		Joins("JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?", userID).
		Where("c.id = ? AND assessments.deleted_at IS NULL", courseID).
		Where("(arl.course_id IS NULL OR arl.course_id = 0 OR arl.course_id = c.id)").
		Where("arl.assigned_by IS NOT NULL AND arl.assigned_by > 0")

	return query
}

func (r *dashboardStudentAssessmentRepository) ListAssessmentsForStudent(
	userID int64,
	courseID int64,
	subjectID *int64,
	assessmentType string,
	assessmentID *int64,
	limit, offset int,
) ([]models.Assessment, int64, error) {
	query := r.baseQuery(userID, courseID)

	if subjectID != nil && *subjectID > 0 {
		query = query.Where("assessments.subject_id = ?", *subjectID)
	}
	if assessmentType != "" {
		query = query.Where("assessments.type = ?", assessmentType)
	}
	if assessmentID != nil && *assessmentID > 0 {
		query = query.Where("assessments.id = ?", *assessmentID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var assessments []models.Assessment
	if err := query.
		Preload("Subject").
		Order("assessments.id DESC").
		Find(&assessments).Error; err != nil {
		return nil, 0, err
	}

	return assessments, total, nil
}
