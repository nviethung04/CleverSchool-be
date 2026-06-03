package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
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
