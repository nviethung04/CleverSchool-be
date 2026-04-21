package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
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
