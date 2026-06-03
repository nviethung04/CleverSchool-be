package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
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
