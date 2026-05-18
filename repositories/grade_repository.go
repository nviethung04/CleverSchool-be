package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type GradeRepository interface {
	base.BaseRepositoryInterface[models.Grade]
}

type gradeRepository struct {
	*base.BaseRepository[models.Grade]
}

func NewGradeRepository() GradeRepository {
	return &gradeRepository{
		BaseRepository: base.NewBaseRepository[models.Grade](),
	}
}
