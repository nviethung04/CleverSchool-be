package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"encoding/json"

	"gorm.io/gorm"
)

type AssessmentScoringRepository interface {
	GetCriteriaGroupWithCriteria(groupID int64) (*models.AssessmentCriteriaGroup, error)
	ListCriteriaGroups(subjectID *int64, keyword string, limit, offset int) ([]models.AssessmentCriteriaGroup, int64, error)
	GetCriteriaGroupByID(id int64) (*models.AssessmentCriteriaGroup, error)
	CreateCriteriaGroup(group *models.AssessmentCriteriaGroup, criteria []models.AssessmentCriterion) error
	UpdateCriteriaGroup(group *models.AssessmentCriteriaGroup, criteria []models.AssessmentCriterion) error
	DeleteCriteriaGroup(id int64, deletedBy int64) error

	GetStudyReportCriteriaByID(id int64) (*models.StudyReportCriteria, error)
	ListStudyReportCriterias(subjectID *int64, keyword string, limit, offset int) ([]models.StudyReportCriteria, int64, error)
	CreateStudyReportCriteria(c *models.StudyReportCriteria) error
	UpdateStudyReportCriteria(c *models.StudyReportCriteria) error
	DeleteStudyReportCriteria(id int64) error

	ListCourseStudentIDs(courseID int64) ([]int64, error)
	GetStudentNames(userIDs []int64) (map[int64]string, error)

	GetAssessmentByID(id int64) (*models.Assessment, error)
	GetScoresByAssessmentCourse(assessmentID, courseID int64) ([]models.AssessmentScore, error)
	GetScore(userID, assessmentID, courseID int64) (*models.AssessmentScore, error)
	UpsertScore(score *models.AssessmentScore, details []models.AssessmentScoreDetail) error

	GetAssessmentPublish(courseID, assessmentID int64) (bool, error)
	SetAssessmentPublish(courseID, assessmentID int64, publish bool) error

	GetStudyReportPublish(courseID, assessmentID int64) (bool, error)
	SetStudyReportPublish(courseID, assessmentID int64, publish bool) error

	ListStudyReports(courseID, assessmentID int64, studentUserID int64) ([]models.StudyReport, error)
	GetStudyReportByID(id int64) (*models.StudyReport, error)
	GetStudyReportByKeys(assessmentID, courseID, studentID int64) (*models.StudyReport, error)
	CreateStudyReport(report *models.StudyReport) error
	UpdateStudyReport(report *models.StudyReport) error
}

type assessmentScoringRepository struct{}

func NewAssessmentScoringRepository() AssessmentScoringRepository {
	return &assessmentScoringRepository{}
}

func (r *assessmentScoringRepository) GetCriteriaGroupWithCriteria(groupID int64) (*models.AssessmentCriteriaGroup, error) {
	var group models.AssessmentCriteriaGroup
	err := db.ReplicaDB.
		Preload("Subject").
		Preload("Criteria", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("sort_order ASC, id ASC")
		}).
		Preload("Criteria.Subcriteria", func(tx *gorm.DB) *gorm.DB {
			return tx.Where("deleted_at IS NULL").Order("sort_order ASC, id ASC")
		}).
		Where("id = ? AND deleted_at IS NULL", groupID).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *assessmentScoringRepository) ListCriteriaGroups(subjectID *int64, keyword string, limit, offset int) ([]models.AssessmentCriteriaGroup, int64, error) {
	q := db.ReplicaDB.Model(&models.AssessmentCriteriaGroup{}).Where("deleted_at IS NULL")
	if subjectID != nil && *subjectID > 0 {
		q = q.Where("subject_id = ?", *subjectID)
	}
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var groups []models.AssessmentCriteriaGroup
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Order("id DESC").Find(&groups).Error; err != nil {
		return nil, 0, err
	}
	return groups, total, nil
}

func (r *assessmentScoringRepository) GetCriteriaGroupByID(id int64) (*models.AssessmentCriteriaGroup, error) {
	return r.GetCriteriaGroupWithCriteria(id)
}

func (r *assessmentScoringRepository) CreateCriteriaGroup(group *models.AssessmentCriteriaGroup, criteria []models.AssessmentCriterion) error {
	return db.MasterDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}
		for i := range criteria {
			criteria[i].GroupId = group.ID
			criteria[i].HasSubcriteria = len(criteria[i].Subcriteria) > 0
			subs := criteria[i].Subcriteria
			criteria[i].Subcriteria = nil
			if err := tx.Create(&criteria[i]).Error; err != nil {
				return err
			}
			for j := range subs {
				subs[j].CriteriaId = criteria[i].ID
				if err := tx.Create(&subs[j]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *assessmentScoringRepository) UpdateCriteriaGroup(group *models.AssessmentCriteriaGroup, criteria []models.AssessmentCriterion) error {
	return db.MasterDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(group).Error; err != nil {
			return err
		}
		// Soft-delete old criteria
		if err := tx.Model(&models.AssessmentCriterion{}).
			Where("group_id = ?", group.ID).
			Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
			return err
		}
		for i := range criteria {
			criteria[i].GroupId = group.ID
			criteria[i].ID = 0
			criteria[i].HasSubcriteria = len(criteria[i].Subcriteria) > 0
			subs := criteria[i].Subcriteria
			criteria[i].Subcriteria = nil
			if err := tx.Create(&criteria[i]).Error; err != nil {
				return err
			}
			for j := range subs {
				subs[j].CriteriaId = criteria[i].ID
				subs[j].ID = 0
				if err := tx.Create(&subs[j]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *assessmentScoringRepository) DeleteCriteriaGroup(id int64, deletedBy int64) error {
	return db.MasterDB.Model(&models.AssessmentCriteriaGroup{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": gorm.Expr("NOW()"), "deleted_by": deletedBy}).Error
}

func (r *assessmentScoringRepository) GetStudyReportCriteriaByID(id int64) (*models.StudyReportCriteria, error) {
	var c models.StudyReportCriteria
	err := db.ReplicaDB.Preload("Subject").Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *assessmentScoringRepository) ListStudyReportCriterias(subjectID *int64, keyword string, limit, offset int) ([]models.StudyReportCriteria, int64, error) {
	q := db.ReplicaDB.Model(&models.StudyReportCriteria{}).Where("deleted_at IS NULL")
	if subjectID != nil && *subjectID > 0 {
		q = q.Where("subject_id = ?", *subjectID)
	}
	if keyword != "" {
		q = q.Where("name ILIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.StudyReportCriteria
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	if err := q.Preload("Subject").Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *assessmentScoringRepository) CreateStudyReportCriteria(c *models.StudyReportCriteria) error {
	return db.MasterDB.Create(c).Error
}

func (r *assessmentScoringRepository) UpdateStudyReportCriteria(c *models.StudyReportCriteria) error {
	return db.MasterDB.Save(c).Error
}

func (r *assessmentScoringRepository) DeleteStudyReportCriteria(id int64) error {
	return db.MasterDB.Model(&models.StudyReportCriteria{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *assessmentScoringRepository) ListCourseStudentIDs(courseID int64) ([]int64, error) {
	var ids []int64
	err := db.ReplicaDB.Table("user_courses uc").
		Select("DISTINCT uc.user_id").
		Joins("JOIN users u ON u.id = uc.user_id AND u.deleted_at IS NULL").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3").
		Where("uc.course_id = ?", courseID).
		Pluck("uc.user_id", &ids).Error
	return ids, err
}

func (r *assessmentScoringRepository) GetStudentNames(userIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(userIDs) == 0 {
		return result, nil
	}
	var users []models.User
	if err := db.ReplicaDB.Select("id, name").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	for _, u := range users {
		result[u.ID] = u.Name
	}
	return result, nil
}

func (r *assessmentScoringRepository) GetAssessmentByID(id int64) (*models.Assessment, error) {
	var a models.Assessment
	err := db.ReplicaDB.Preload("Subject").Where("id = ? AND deleted_at IS NULL", id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *assessmentScoringRepository) GetScoresByAssessmentCourse(assessmentID, courseID int64) ([]models.AssessmentScore, error) {
	var scores []models.AssessmentScore
	err := db.ReplicaDB.
		Preload("Details").
		Where("assessment_id = ? AND course_id = ?", assessmentID, courseID).
		Find(&scores).Error
	return scores, err
}

func (r *assessmentScoringRepository) GetScore(userID, assessmentID, courseID int64) (*models.AssessmentScore, error) {
	var score models.AssessmentScore
	err := db.ReplicaDB.
		Preload("Details").
		Where("user_id = ? AND assessment_id = ? AND course_id = ?", userID, assessmentID, courseID).
		First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

func (r *assessmentScoringRepository) UpsertScore(score *models.AssessmentScore, details []models.AssessmentScoreDetail) error {
	return db.MasterDB.Transaction(func(tx *gorm.DB) error {
		var existing models.AssessmentScore
		err := tx.Where("user_id = ? AND assessment_id = ? AND course_id = ?",
			score.UserId, score.AssessmentId, score.CourseId).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := tx.Create(score).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			score.ID = existing.ID
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"total_score": score.TotalScore,
				"file_infos":  score.FileInfos,
				"is_scored":   score.IsScored,
				"updated_by":  score.UpdatedBy,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("assessment_score_id = ?", score.ID).Delete(&models.AssessmentScoreDetail{}).Error; err != nil {
				return err
			}
		}
		for i := range details {
			details[i].AssessmentScoreId = score.ID
			if err := tx.Create(&details[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *assessmentScoringRepository) GetAssessmentPublish(courseID, assessmentID int64) (bool, error) {
	var pub models.AssessmentPublish
	err := db.ReplicaDB.Where("course_id = ? AND assessment_id = ?", courseID, assessmentID).First(&pub).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return pub.Publish, nil
}

func (r *assessmentScoringRepository) SetAssessmentPublish(courseID, assessmentID int64, publish bool) error {
	pub := models.AssessmentPublish{CourseId: courseID, AssessmentId: assessmentID, Publish: publish}
	return db.MasterDB.Save(&pub).Error
}

func (r *assessmentScoringRepository) GetStudyReportPublish(courseID, assessmentID int64) (bool, error) {
	var pub models.StudyReportPublish
	err := db.ReplicaDB.Where("course_id = ? AND assessment_id = ?", courseID, assessmentID).First(&pub).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return pub.Publish, nil
}

func (r *assessmentScoringRepository) SetStudyReportPublish(courseID, assessmentID int64, publish bool) error {
	pub := models.StudyReportPublish{CourseId: courseID, AssessmentId: assessmentID, Publish: publish}
	return db.MasterDB.Save(&pub).Error
}

func (r *assessmentScoringRepository) ListStudyReports(courseID, assessmentID int64, studentUserID int64) ([]models.StudyReport, error) {
	q := db.ReplicaDB.Preload("Assessment").Preload("Student").Preload("Teacher").Preload("Course").
		Where("course_id = ?", courseID)
	if assessmentID > 0 {
		q = q.Where("assessment_id = ?", assessmentID)
	}
	if studentUserID > 0 {
		q = q.Where("student_id = ?", studentUserID)
	}
	var list []models.StudyReport
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *assessmentScoringRepository) GetStudyReportByID(id int64) (*models.StudyReport, error) {
	var report models.StudyReport
	err := db.ReplicaDB.
		Preload("Assessment").
		Preload("Student").
		Preload("Teacher").
		Preload("Course").
		First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *assessmentScoringRepository) GetStudyReportByKeys(assessmentID, courseID, studentID int64) (*models.StudyReport, error) {
	var report models.StudyReport
	err := db.ReplicaDB.Where("assessment_id = ? AND course_id = ? AND student_id = ?",
		assessmentID, courseID, studentID).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *assessmentScoringRepository) CreateStudyReport(report *models.StudyReport) error {
	return db.MasterDB.Create(report).Error
}

func (r *assessmentScoringRepository) UpdateStudyReport(report *models.StudyReport) error {
	return db.MasterDB.Save(report).Error
}

func MarshalJSONRaw(v interface{}) models.JSONRaw {
	b, _ := json.Marshal(v)
	return models.JSONRaw(b)
}

func ParseJSONRaw(raw models.JSONRaw, dest interface{}) {
	if len(raw) == 0 {
		return
	}
	_ = json.Unmarshal([]byte(raw), dest)
}
