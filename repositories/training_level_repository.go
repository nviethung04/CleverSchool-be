package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
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
