package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type TopicController struct {
	*GenericController[models.Topic, prot.Topic, *prot.TopicRequest]
}

func NewTopicController(service services.TopicService) *TopicController {
	topicResource := resources.NewTopicResource()
	topicResourceAdapter := NewTopicResourceAdapter(topicResource)

	genericController := NewGenericController(
		service,
		topicResourceAdapter,
		func() *prot.TopicRequest {
			return &prot.TopicRequest{}
		},
		func(topics []*prot.Topic, totalCount uint64) interface{} {
			return &prot.TopicsResponse{
				Topics:     topics,
				TotalCount: totalCount,
			}
		},
	)

	return &TopicController{
		GenericController: genericController,
	}
}
