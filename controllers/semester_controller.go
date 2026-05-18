package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type SemesterController struct {
	*GenericController[models.Semester, prot.Semester, *prot.SemesterRequest]
	service services.SemesterService
}

func NewSemesterController(service services.SemesterService) *SemesterController {
	semesterResource := resources.NewSemesterResource()
	semesterResourceAdapter := NewSemesterResourceAdapter(semesterResource)

	genericController := NewGenericController(
		service,
		semesterResourceAdapter,
		func() *prot.SemesterRequest {
			return &prot.SemesterRequest{}
		},
		func(semesters []*prot.Semester, totalCount uint64) interface{} {
			return &prot.SemestersResponse{
				Semesters:  semesters,
				TotalCount: int64(totalCount),
			}
		},
	)

	ctl := &SemesterController{
		GenericController: genericController,
		service:           service,
	}

	return ctl
}
