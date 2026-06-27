package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AssessmentController struct {
	*GenericController[models.Assessment, prot.Assessment, *prot.AssessmentRequest]
	svc services.AssessmentService
}

func NewAssessmentController(service services.AssessmentService) *AssessmentController {
	assessmentResource := resources.NewAssessmentResource()
	assessmentResourceAdapter := NewAssessmentResourceAdapter(assessmentResource)

	genericController := NewGenericController(
		service,
		assessmentResourceAdapter,
		func() *prot.AssessmentRequest {
			return &prot.AssessmentRequest{}
		},
		func(assessments []*prot.Assessment, totalCount uint64) interface{} {
			return &prot.AssessmentsResponse{
				Assessments: assessments,
				Total:       int64(totalCount),
			}
		},
	)

	return &AssessmentController{
		GenericController: genericController,
		svc:               service,
	}
}

func (ctl *AssessmentController) Assigned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	assigned, err := ctl.svc.Assigned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}
	utils.Respond(c, assigned, err, "")
}

func (ctl *AssessmentController) AssignedLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	lessons, err := ctl.svc.AssignedLesson(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.no_records_found", http.StatusNotFound)
		return
	}
	utils.Respond(c, lessons, err, "")
}
