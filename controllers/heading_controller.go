package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
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

