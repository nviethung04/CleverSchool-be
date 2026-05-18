package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type HeadingController struct {
	*GenericController[models.Heading, prot.Heading, *prot.HeadingRequest]
}

func NewHeadingController(service services.HeadingService) *HeadingController {
	headingResource := resources.NewHeadingResource()
	headingResourceAdapter := NewHeadingResourceAdapter(headingResource)

	genericController := NewGenericController(
		service,
		headingResourceAdapter,
		func() *prot.HeadingRequest {
			return &prot.HeadingRequest{}
		},
		func(headings []*prot.Heading, totalCount uint64) interface{} {
			return &prot.HeadingsResponse{
				Headings:   headings,
				TotalCount: totalCount,
			}
		},
	)

	ctl := &HeadingController{
		GenericController: genericController,
	}

	ctl.GenericController.WithUsedError(func() bool {
		return true
	})

	return ctl
}
