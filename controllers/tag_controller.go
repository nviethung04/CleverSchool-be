package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
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

