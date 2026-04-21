package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type StudyReportCriteriaController struct {
	*GenericController[models.StudyReportCriteria, prot.StudyReportCriteria, *prot.StudyReportCriteriaRequest]
}

func NewStudyReportCriteriaController(service services.StudyReportCriteriaService) *StudyReportCriteriaController {
	studyReportCriteriaResource := resources.NewStudyReportCriteriaResource()
	studyReportCriteriaResourceAdapter := NewStudyReportCriteriaResourceAdapter(studyReportCriteriaResource)

	genericController := NewGenericController(
		service,
		studyReportCriteriaResourceAdapter,
		func() *prot.StudyReportCriteriaRequest {
			return &prot.StudyReportCriteriaRequest{}
		},
		func(studyReportCriterias []*prot.StudyReportCriteria, totalCount uint64) interface{} {
			return &prot.StudyReportCriteriasResponse{
				StudyReportCriterias: studyReportCriterias,
				TotalCount:           totalCount,
			}
		},
	)

	return &StudyReportCriteriaController{
		GenericController: genericController,
	}
}
