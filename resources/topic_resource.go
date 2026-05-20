package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
)

type TopicResource interface {
	FormatTopic(topic *models.Topic) *prot.Topic
	FormatTopics(topics []*models.Topic) []*prot.Topic
	FormatModelTopic(topic *prot.TopicRequest) *models.Topic
}

type TopicResourceImpl struct{}

func NewTopicResource() TopicResource {
	return &TopicResourceImpl{}
}

func (r *TopicResourceImpl) FormatTopic(topic *models.Topic) *prot.Topic {
	if topic == nil {
		return nil
	}

	return &prot.Topic{
		Id:          int64(topic.ID),
		Name:        topic.Name,
		Type:        topic.Type,
		ParentId:    utils.Int64OrZero(topic.ParentID),
		Description: topic.Description,
		ImageUrl:    utils.StaticURL(topic.ImageInfo.Path, models.Storage),
		Status:      topic.Status,
		CreatedAt:   topic.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   topic.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *TopicResourceImpl) FormatTopics(topics []*models.Topic) []*prot.Topic {
	result := make([]*prot.Topic, 0, len(topics))
	for _, t := range topics {
		if formatted := r.FormatTopic(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *TopicResourceImpl) FormatModelTopic(topic *prot.TopicRequest) *models.Topic {
	if topic == nil {
		return nil
	}

	imageUrl := utils.StripDomain(topic.ImageUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(imageUrl, models.Storage)

	return &models.Topic{
		ID:          int64(topic.Id),
		Name:        topic.Name,
		Type:        topic.Type,
		ParentID:    utils.Int64PtrOrNil(topic.ParentId),
		Description: topic.Description,
		ImageInfo:   imageInfo,
		Status:      topic.Status,
	}
}
