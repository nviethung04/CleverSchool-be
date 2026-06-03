package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type ProgramController struct {
	*GenericController[models.Program, prot.Program, *prot.ProgramRequest]
	service services.ProgramService
}

func NewProgramController(service services.ProgramService) *ProgramController {
	programResource := resources.NewProgramResource()
	programResourceAdapter := NewProgramResourceAdapter(programResource)

	genericController := NewGenericController(
		service,
		programResourceAdapter,
		func() *prot.ProgramRequest {
			return &prot.ProgramRequest{}
		},
		func(programs []*prot.Program, totalCount uint64) interface{} {
			return &prot.ProgramsResponse{
				Programs:   programs,
				TotalCount: totalCount,
			}
		},
	)

	return &ProgramController{
		GenericController: genericController,
		service:           service,
	}
}
