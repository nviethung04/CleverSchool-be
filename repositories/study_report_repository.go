package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StudyReportSkillValue struct {
	SkillID     int64
	SkillTypeID int64
	StarValue   *int32
	CheckValue  *bool
}

type StudyReportEvaluateData struct {
	Students []models.User
	Course   models.Course
	Reports  []models.StudyReport
}

type StudyReportRepository interface {
	base.BaseRepositoryInterface[models.StudyReport]
	ReplaceSkillValues(reportID, criteriaID int64, values []StudyReportSkillValue) error
	ValidateStudyReportIDs(studentId, assessmentId, courseId int64) (int64, error)
	GetEvaluateData(courseId, assessmentId int64) (*StudyReportEvaluateData, error)
	FindByStudentCourseSubjectAssessment(studentId, courseId, subjectId, assessmentId int64) (*models.StudyReport, error)

	GetPublish(courseId, assessmentId int64) bool
	UpdatePublish(courseId, assessmentId int64, status bool) bool
	FindPublishReports(courseIds, assessmentIds []int64) ([]models.PublishReport, error)

	UpdateIsCompleted(id int64, isCompleted bool) bool
	UpdateTotalStarAndAvgStar(id int64, totalStar int, avgStar float64) error
}

type studyReportRepository struct {
	*base.BaseRepository[models.StudyReport]
}

func NewStudyReportRepository() StudyReportRepository {
	return &studyReportRepository{
		BaseRepository: base.NewBaseRepository[models.StudyReport](),
	}
}

func (r *studyReportRepository) ReplaceSkillValues(reportID, criteriaID int64, values []StudyReportSkillValue) error {
	tx := db.MasterDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("study_report_id = ?", reportID).
		Delete(&models.StudyReportRefSkillType{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, val := range values {
		entry := models.StudyReportRefSkillType{
			StudyReportId:         reportID,
			StudyReportCriteriaId: criteriaID,
			SkillId:               val.SkillID,
			SkillTypeId:           val.SkillTypeID,
		}

		if val.StarValue != nil {
			star := int(*val.StarValue)
			entry.Star = &star
		}

		if val.CheckValue != nil {
			entry.IsCheck = *val.CheckValue
		}

		if err := tx.Create(&entry).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *studyReportRepository) ValidateStudyReportIDs(
	studentId, assessmentId, courseId int64,
) (int64, error) {
	// Nếu courseId > 0, validate courseId
	if courseId > 0 {
		var count int64
		err := db.ReplicaDB.Raw(`
			SELECT COUNT(*)
			FROM courses c
			WHERE c.id = ?
			LIMIT 1;
		`, courseId).Scan(&count).Error

		if err != nil {
			return 0, err
		}

		if count == 0 {
			return 0, fmt.Errorf("Course %d does not exist", courseId)
		}

		return courseId, nil
	}

	// Nếu courseId == 0, tìm courseId theo studentId và assessmentId
	var res struct {
		CourseID *int64 `gorm:"column:course_id"`
	}

	err := db.ReplicaDB.Raw(`
		SELECT ar.course_id
		FROM assessment_ref_lessons ar
		JOIN user_courses uc ON uc.course_id = ar.course_id
		WHERE uc.user_id = ?
		  AND ar.assessment_id = ?
		LIMIT 1;
	`, studentId, assessmentId).Scan(&res).Error

	if err != nil {
		return 0, err
	}

	if res.CourseID == nil || *res.CourseID == 0 {
		return 0, fmt.Errorf(
			"Student %d is not enrolled in the course of assessment %d",
			studentId, assessmentId,
		)
	}

	return *res.CourseID, nil
}

func (r *studyReportRepository) GetEvaluateData(courseId, assessmentId int64) (*StudyReportEvaluateData, error) {
	var course models.Course
	if err := db.ReplicaDB.Where("id = ?", courseId).First(&course).Error; err != nil {
		return nil, err
	}

	courseRepo := NewCourseRepository()
	students, _, err := courseRepo.GetUsers(courseId, models.StudentRoleId)
	if err != nil {
		return nil, err
	}

	var reports []models.StudyReport
	if err := db.ReplicaDB.
		Where("course_id = ? AND assessment_id = ?", courseId, assessmentId).
		Preload("Criteria").
		Preload("Criteria.Skills").
		Preload("Criteria.Skills.Types").
		Preload("Skills").
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return &StudyReportEvaluateData{
		Students: students,
		Course:   course,
		Reports:  reports,
	}, nil
}

func (r *studyReportRepository) FindByStudentCourseSubjectAssessment(studentId, courseId, subjectId, assessmentId int64) (*models.StudyReport, error) {
	var report models.StudyReport
	err := db.ReplicaDB.Where("student_id = ? AND course_id = ? AND subject_id = ? AND assessment_id = ?",
		studentId, courseId, subjectId, assessmentId).
		First(&report).Error

	if err != nil {
		return nil, err
	}

	return &report, nil
}

func (r *studyReportRepository) FindPublishReports(courseIds, assessmentIds []int64) ([]models.PublishReport, error) {
	publishes := make([]models.PublishReport, 0)

	if len(courseIds) == 0 || len(assessmentIds) == 0 {
		return publishes, nil
	}

	if err := db.ReplicaDB.
		Where("course_id IN ? AND assessment_id IN ?", courseIds, assessmentIds).
		Find(&publishes).Error; err != nil {
		return nil, err
	}

	return publishes, nil
}

func (r *studyReportRepository) GetPublish(courseId, assessmentId int64) bool {
	var publish models.PublishReport

	err := db.ReplicaDB.
		Where("course_id = ? AND assessment_id = ?", courseId, assessmentId).
		First(&publish).Error

	if err != nil || err == gorm.ErrRecordNotFound {
		return false
	}

	return true
}

func (r *studyReportRepository) UpdatePublish(courseId, assessmentId int64, status bool) bool {
	var publish models.PublishReport

	err := db.MasterDB.
		Where("course_id = ? AND assessment_id = ?", courseId, assessmentId).
		First(&publish).Error

	if status == false {
		if err == nil {
			if err := db.MasterDB.
				Where("course_id = ? AND assessment_id = ?", courseId, assessmentId).
				Delete(&models.PublishReport{}).Error; err != nil {
				return false
			}
		}
		return true
	}

	if err == nil {
		return true
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newPublish := models.PublishReport{
			CourseId:     courseId,
			AssessmentId: assessmentId,
		}

		if err := db.MasterDB.Create(&newPublish).Error; err != nil {
			return false
		}
		return true
	}

	return false
}

func (r *studyReportRepository) UpdateIsCompleted(id int64, isCompleted bool) bool {
	if err := db.MasterDB.Model(&models.StudyReport{}).
		Where("id = ?", id).
		Update("is_completed", isCompleted).Error; err != nil {
		return false
	}

	return true
}

func (r *studyReportRepository) UpdateTotalStarAndAvgStar(id int64, totalStar int, avgStar float64) error {
	return db.MasterDB.Model(&models.StudyReport{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_star": totalStar,
			"avg_star":   avgStar,
		}).Error
}
