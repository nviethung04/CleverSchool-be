package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type TagController struct {
	*GenericController[models.Tag, prot.Tag, *prot.TagRequest]
}

func NewTagController(service services.TagService) *TagController {
	tagResource := resources.NewTagResource()
	tagResourceAdapter := NewTagResourceAdapter(tagResource)

	genericController := NewGenericController(
		service,
		tagResourceAdapter,
		func() *prot.TagRequest {
			return &prot.TagRequest{}
		},
		func(tags []*prot.Tag, totalCount uint64) interface{} {
			return &prot.TagsResponse{
				Tags:       tags,
				TotalCount: totalCount,
			}
		},
	)

	return &TagController{
		GenericController: genericController,
	}
}
