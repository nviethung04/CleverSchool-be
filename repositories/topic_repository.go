package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type TopicRepository interface {
	base.BaseRepositoryInterface[models.Topic]
}

type topicRepository struct {
	*base.BaseRepository[models.Topic]
}

func NewTopicRepository() TopicRepository {
	return &topicRepository{
		BaseRepository: base.NewBaseRepository[models.Topic](),
	}
}
