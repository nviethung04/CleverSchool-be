package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
)

type SubjectRepository interface {
	base.BaseRepositoryInterface[models.Subject]
}

type subjectRepository struct {
	*base.BaseRepository[models.Subject]
}

func NewSubjectRepository() SubjectRepository {
	return &subjectRepository{
		BaseRepository: base.NewBaseRepository[models.Subject](),
	}
}
