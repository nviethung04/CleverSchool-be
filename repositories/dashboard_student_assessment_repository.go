package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
)

type DashboardStudentAssessmentRepository interface {
	GetAssessmentsByCourseID(courseID int64, assessmentID *int64, assessmentType *string) ([]models.AssessmentRefLesson, error)
}

type dashboardStudentAssessmentRepository struct{}

func NewDashboardStudentAssessmentRepository() DashboardStudentAssessmentRepository {
	return &dashboardStudentAssessmentRepository{}
}

func (r *dashboardStudentAssessmentRepository) GetAssessmentsByCourseID(courseID int64, assessmentID *int64, assessmentType *string) ([]models.AssessmentRefLesson, error) {
	var refs []models.AssessmentRefLesson
	query := db.ReplicaDB.
		Table("assessment_ref_lessons").
		Select("assessment_ref_lessons.*").
		Joins("INNER JOIN assessments ON assessments.id = assessment_ref_lessons.assessment_id").
		Where("assessment_ref_lessons.course_id = ? AND assessment_ref_lessons.assigned_at IS NOT NULL", courseID)
	
	if assessmentID != nil && *assessmentID > 0 {
		query = query.Where("assessment_ref_lessons.assessment_id = ?", *assessmentID)
	}
	
	if assessmentType != nil && *assessmentType != "" {
		query = query.Where("assessments.type = ?", *assessmentType)
	}
	
	err := query.
		Order("assessment_ref_lessons.assigned_at DESC").
		Find(&refs).Error
	return refs, err
}

