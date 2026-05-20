package repositories

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
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

