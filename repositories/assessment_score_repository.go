package repositories

import (
	"time"

	"be-cleverschool/database/db"
	"be-cleverschool/models"
)

type AssessmentScoreRepository interface {
	Create(score *models.AssessmentScore) error
	CreateDetail(detail *models.AssessmentScoreDetail) error
	CreateDetails(details []*models.AssessmentScoreDetail) error
	GetScoreByStudentAndAssessment(studentID, assessmentID, courseID int64) (*models.AssessmentScore, error)
	GetScoreByStudentAndAssessmentWithoutCourse(studentID, assessmentID int64) (*models.AssessmentScore, error)
	GetScoresByAssessmentAndStudents(assessmentID int64, studentIDs []int64) ([]models.AssessmentScore, error)
	GetScoreDetailsByScoreID(scoreID int64) ([]models.AssessmentScoreDetail, error)
	DeleteOldScores(assessmentID, studentID, courseID int64, deletedBy int64) error
}

type assessmentScoreRepository struct{}

func NewAssessmentScoreRepository() AssessmentScoreRepository {
	return &assessmentScoreRepository{}
}

func (r *assessmentScoreRepository) Create(score *models.AssessmentScore) error {
	return db.MasterDB.Create(score).Error
}

func (r *assessmentScoreRepository) CreateDetail(detail *models.AssessmentScoreDetail) error {
	return db.MasterDB.Create(detail).Error
}

func (r *assessmentScoreRepository) CreateDetails(details []*models.AssessmentScoreDetail) error {
	if len(details) == 0 {
		return nil
	}
	return db.MasterDB.Create(details).Error
}

func (r *assessmentScoreRepository) GetScoreByStudentAndAssessment(studentID, assessmentID, courseID int64) (*models.AssessmentScore, error) {
	var score models.AssessmentScore
	err := db.ReplicaDB.
		Where("student_id = ? AND assessment_id = ? AND course_id = ? AND deleted_at IS NULL", studentID, assessmentID, courseID).
		Order("created_at DESC").
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (r *assessmentScoreRepository) GetScoreByStudentAndAssessmentWithoutCourse(studentID, assessmentID int64) (*models.AssessmentScore, error) {
	var score models.AssessmentScore
	err := db.ReplicaDB.
		Where("student_id = ? AND assessment_id = ? AND deleted_at IS NULL", studentID, assessmentID).
		Order("created_at DESC").
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// GetScoresByAssessmentAndStudents lấy tất cả bản ghi điểm của 1 assessment cho 1 list student_id
// (API report sẽ xử lý chọn bản ghi mới nhất cho từng student ở tầng service)
func (r *assessmentScoreRepository) GetScoresByAssessmentAndStudents(assessmentID int64, studentIDs []int64) ([]models.AssessmentScore, error) {
	var scores []models.AssessmentScore
	if len(studentIDs) == 0 {
		return scores, nil
	}

	err := db.ReplicaDB.
		Where("assessment_id = ? AND student_id IN (?) AND deleted_at IS NULL", assessmentID, studentIDs).
		Order("student_id, created_at DESC").
		Find(&scores).Error
	if err != nil {
		return nil, err
	}

	return scores, nil
}

func (r *assessmentScoreRepository) GetScoreDetailsByScoreID(scoreID int64) ([]models.AssessmentScoreDetail, error) {
	var details []models.AssessmentScoreDetail
	err := db.ReplicaDB.
		Where("assessment_score_id = ? AND deleted_at IS NULL", scoreID).
		Find(&details).Error
	return details, err
}

func (r *assessmentScoreRepository) DeleteOldScores(assessmentID, studentID, courseID int64, deletedBy int64) error {
	now := time.Now().UTC()

	// Lấy danh sách score IDs cần xóa
	var scoreIDs []int64
	err := db.MasterDB.Model(&models.AssessmentScore{}).
		Where("assessment_id = ? AND student_id = ? AND course_id = ? AND deleted_at IS NULL", assessmentID, studentID, courseID).
		Pluck("id", &scoreIDs).Error
	if err != nil {
		return err
	}

	if len(scoreIDs) == 0 {
		return nil
	}

	// Soft delete assessment_score_details trước
	err = db.MasterDB.Model(&models.AssessmentScoreDetail{}).
		Where("assessment_score_id IN (?) AND deleted_at IS NULL", scoreIDs).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
			"updated_by": deletedBy,
		}).Error
	if err != nil {
		return err
	}

	// Soft delete assessment_scores
	err = db.MasterDB.Model(&models.AssessmentScore{}).
		Where("id IN (?)", scoreIDs).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": deletedBy,
			"updated_at": now,
			"updated_by": deletedBy,
		}).Error
	if err != nil {
		return err
	}

	return nil
}

