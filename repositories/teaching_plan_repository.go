package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
)

type TeachingPlanRepository interface {
	base.BaseRepositoryInterface[models.TeachingPlan]
	UpdateLessons(id int64, lessonIds []int64) error
}

type teachingPlanRepository struct {
	*base.BaseRepository[models.TeachingPlan]
}

func NewTeachingPlanRepository() TeachingPlanRepository {
	return &teachingPlanRepository{
		BaseRepository: base.NewBaseRepository[models.TeachingPlan](),
	}
}

func (r *teachingPlanRepository) UpdateLessons(id int64, lessonIds []int64) error {
	if id <= 0 {
		return nil
	}

	if err := db.MasterDB.
		Where("teaching_plan_id = ?", id).
		Delete(&models.LessonRefTeachingPlan{}).Error; err != nil {
		return err
	}

	if len(lessonIds) == 0 {
		return nil
	}

	refs := make([]models.LessonRefTeachingPlan, 0, len(lessonIds))
	for _, lessonId := range lessonIds {
		if lessonId <= 0 {
			continue
		}
		refs = append(refs, models.LessonRefTeachingPlan{
			LessonId:       lessonId,
			TeachingPlanId: id,
		})
	}

	if len(refs) > 0 {
		if err := db.MasterDB.Create(&refs).Error; err != nil {
			return err
		}
	}

	return nil
}

