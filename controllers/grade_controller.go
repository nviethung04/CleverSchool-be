package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
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
