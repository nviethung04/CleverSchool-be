package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudyReportController struct {
	*GenericController[models.StudyReport, prot.StudyReport, *prot.StudyReportRequest]
	svc services.StudyReportService
}

func NewStudyReportController(service services.StudyReportService) *StudyReportController {
	studyReportResource := resources.NewStudyReportResource()
	studyReportResourceAdapter := NewStudyReportResourceAdapter(studyReportResource)

	genericController := NewGenericController(
		service,
		studyReportResourceAdapter,
		func() *prot.StudyReportRequest {
			return &prot.StudyReportRequest{}
		},
		func(studyReports []*prot.StudyReport, totalCount uint64) interface{} {
			return &prot.StudyReportsResponse{
				StudyReports: studyReports,
				TotalCount:   totalCount,
			}
		},
	)

	ctl := &StudyReportController{
		GenericController: genericController,
		svc:               service,
	}

	ctl.GenericController.WithUsedError(func() bool {
		return true
	})

	return ctl
}

func (ctl *StudyReportController) Evaluate(c *gin.Context) {
	result, err := ctl.svc.Evaluate(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, result, nil, "")
}

func (ctl *StudyReportController) GetPublish(c *gin.Context) {
	result, err := ctl.svc.GetPublish(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, result, nil, "")
}

func (ctl *StudyReportController) UpdatePublish(c *gin.Context) {
	result, err := ctl.svc.UpdatePublish(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, result, nil, "")
}
