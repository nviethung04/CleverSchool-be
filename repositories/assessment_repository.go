package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"fmt"
)

type AssessmentRepository interface {
	base.BaseRepositoryInterface[models.Assessment]
	Assigned(ref models.AssessmentRefLesson) error
	AssignedLesson(id int64) (*models.Assessment, error)
}

type assessmentRepository struct {
	*base.BaseRepository[models.Assessment]
}

func NewAssessmentRepository() AssessmentRepository {
	return &assessmentRepository{
		BaseRepository: base.NewBaseRepository[models.Assessment](),
	}
}

func assessmentOmitZeroFK(entity *models.Assessment) []string {
	omit := []string{"author_id"}
	if entity.ProgramId == 0 {
		omit = append(omit, "ProgramId")
	}
	if entity.SubjectId == 0 {
		omit = append(omit, "SubjectId")
	}
	if entity.AssessmentCriteriaGroupId == 0 {
		omit = append(omit, "AssessmentCriteriaGroupId")
	}
	if entity.StudyReportCriteriaId == 0 {
		omit = append(omit, "StudyReportCriteriaId")
	}
	return omit
}

func (r *assessmentRepository) Create(entity *models.Assessment) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	if err := r.BeforeCreate(entity); err != nil {
		return err
	}
	return db.MasterDB.Omit(assessmentOmitZeroFK(entity)...).Create(entity).Error
}

func (r *assessmentRepository) Update(entity *models.Assessment) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	if err := r.BeforeUpdate(entity); err != nil {
		return err
	}
	omit := append([]string{"created_at", "created_by"}, assessmentOmitZeroFK(entity)...)
	return db.MasterDB.Omit(omit...).Save(entity).Error
}

// Override BaseRepository.Delete to avoid Save() overwriting nullable FK fields (program_id,
// subject_id, ...) with 0, which violates FK constraints like assessments_program_id_fkey.
func (r *assessmentRepository) Delete(id int) error {
	var model models.Assessment
	if err := r.BeforeDelete(id, &model); err != nil {
		return err
	}
	if err := db.MasterDB.Model(&models.Assessment{}).
		Where("id = ?", id).
		Update("deleted_by", model.DeletedBy).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&models.Assessment{}, id).Error
}
