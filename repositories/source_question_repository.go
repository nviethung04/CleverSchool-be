package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
)

type SourceQuestionRepository interface {
	base.BaseRepositoryInterface[models.SourceQuestion]
}

type sourceQuestionRepository struct {
	*base.BaseRepository[models.SourceQuestion]
}

func NewSourceQuestionRepository() SourceQuestionRepository {
	return &sourceQuestionRepository{
		BaseRepository: base.NewBaseRepository[models.SourceQuestion](),
	}
}
