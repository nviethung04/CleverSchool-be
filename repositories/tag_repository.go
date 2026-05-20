package repositories

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
)

type TagRepository interface {
	base.BaseRepositoryInterface[models.Tag]
}

type tagRepository struct {
	*base.BaseRepository[models.Tag]
}

func NewTagRepository() TagRepository {
	return &tagRepository{
		BaseRepository: base.NewBaseRepository[models.Tag](),
	}
}

