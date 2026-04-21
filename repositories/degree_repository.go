package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
)

type DegreeRepository interface {
	base.BaseRepositoryInterface[models.Degree]
}

type degreeRepository struct {
	*base.BaseRepository[models.Degree]
}

func NewDegreeRepository() DegreeRepository {
	return &degreeRepository{
		BaseRepository: base.NewBaseRepository[models.Degree](),
	}
}
