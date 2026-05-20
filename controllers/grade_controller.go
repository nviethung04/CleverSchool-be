package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type GradeController struct {
	*GenericController[models.Grade, prot.Grade, *prot.GradeRequest]
}

func NewGradeController(service services.GradeService) *GradeController {
	gradeResource := resources.NewGradeResource()
	gradeResourceAdapter := NewGradeResourceAdapter(gradeResource)

	genericController := NewGenericController(
		service,
		gradeResourceAdapter,
		func() *prot.GradeRequest {
			return &prot.GradeRequest{}
		},
		func(grades []*prot.Grade, totalCount uint64) interface{} {
			return &prot.GradesResponse{
				Grades:     grades,
				TotalCount: totalCount,
			}
		},
	)

	return &GradeController{
		GenericController: genericController,
	}
}
