package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type StudyShiftController struct {
	*GenericController[models.StudyShift, prot.StudyShift, *prot.StudyShiftRequest]
}

func NewStudyShiftController(service services.StudyShiftService) *StudyShiftController {
	studyShiftResource := resources.NewStudyShiftResource()
	studyShiftResourceAdapter := NewStudyShiftResourceAdapter(studyShiftResource)

	genericController := NewGenericController(
		service,
		studyShiftResourceAdapter,
		func() *prot.StudyShiftRequest {
			return &prot.StudyShiftRequest{}
		},
		func(studyShifts []*prot.StudyShift, totalCount uint64) interface{} {
			return &prot.StudyShiftsResponse{
				StudyShifts:       studyShifts,
				TotalCount: totalCount,
			}
		},
	)

	return &StudyShiftController{
		GenericController: genericController,
	}
}

