package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type DegreeController struct {
	*GenericController[models.Degree, prot.Degree, *prot.Degree]
}

func NewDegreeController(service services.DegreeService) *DegreeController {
	degreeResource := resources.NewDegreeResource()
	degreeResourceAdapter := NewDegreeResourceAdapter(degreeResource)

	genericController := NewGenericController(
		service,
		degreeResourceAdapter,
		func() *prot.Degree {
			return &prot.Degree{}
		},
		func(degrees []*prot.Degree, totalCount uint64) interface{} {
			return &prot.DegreelListResponse{
				Degrees:       degrees,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &DegreeController{
		GenericController: genericController,
	}
}
