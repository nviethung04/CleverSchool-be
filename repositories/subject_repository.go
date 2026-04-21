package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SubjectRepository interface {
	base.BaseRepositoryInterface[models.Subject]
	UpdateTrainingLevels(id int64, trainingLevelIds []int64) error
}

type subjectRepository struct {
	*base.BaseRepository[models.Subject]
}

func NewSubjectRepository() SubjectRepository {
	repo := &subjectRepository{
		BaseRepository: base.NewBaseRepository[models.Subject](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *subjectRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	schoolId := r.GetAdminSchoolId(ctx)

	if schoolId <= 0 {
		return query
	}

	ids, err := r.GetSubjectIDsBySchool(int64(schoolId))
	if err != nil || len(ids) == 0 {
		return query.Where("1 = 0")
	}

	return query.Where("id IN ?", ids)
}

func (r *subjectRepository) GetSubjectIDsBySchool(schoolId int64) ([]int64, error) {
	if schoolId <= 0 {
		return nil, nil
	}

	var subjectIDs []int64

	err := db.ReplicaDB.
		Table("subjects").
		Select("subjects.id").
		Joins("JOIN programs ON programs.subject_id = subjects.id").
		Joins("JOIN courses ON courses.program_id = programs.id").
		Joins("JOIN course_schools ON course_schools.course_id = courses.id").
		Where("course_schools.school_id = ?", schoolId).
		Distinct().
		Pluck("subjects.id", &subjectIDs).Error

	if err != nil {
		return nil, err
	}

	return subjectIDs, nil
}

func (r *subjectRepository) UpdateTrainingLevels(subjectID int64, trainingLevelIDs []int64) error {
	if subjectID <= 0 {
		return nil
	}

	if err := db.MasterDB.
		Where("subject_id = ?", subjectID).
		Delete(&models.SubjectRefTrainingLevel{}).Error; err != nil {
		return err
	}

	if len(trainingLevelIDs) == 0 {
		return nil
	}

	refs := make([]models.SubjectRefTrainingLevel, 0, len(trainingLevelIDs))
	for _, levelID := range trainingLevelIDs {
		if levelID <= 0 {
			continue
		}
		refs = append(refs, models.SubjectRefTrainingLevel{
			SubjectID:       subjectID,
			TrainingLevelID: levelID,
		})
	}

	if len(refs) > 0 {
		if err := db.MasterDB.Create(&refs).Error; err != nil {
			return err
		}
	}

	return nil
}
