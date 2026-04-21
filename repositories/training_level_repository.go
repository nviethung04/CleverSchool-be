package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
)

type TrainingLevelRepository interface {
	base.BaseRepositoryInterface[models.TrainingLevel]
}

type trainingLevelRepository struct {
	*base.BaseRepository[models.TrainingLevel]
}

func NewTrainingLevelRepository() TrainingLevelRepository {
	return &trainingLevelRepository{
		BaseRepository: base.NewBaseRepository[models.TrainingLevel](),
	}
}
