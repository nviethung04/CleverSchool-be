package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssessmentController struct {
	svc services.AssessmentService
	*GenericController[models.Assessment, prot.Assessment, *prot.AssessmentRequest]
}

func NewAssessmentController(service services.AssessmentService) *AssessmentController {
	resource := resources.NewAssessmentResource()
	adapter := NewAssessmentResourceAdapter(resource)

	genericController := NewGenericController(
		service,
		adapter,
		func() *prot.AssessmentRequest {
			return &prot.AssessmentRequest{}
		},
		func(items []*prot.Assessment, totalCount uint64) interface{} {
			return &prot.AssessmentListResponse{
				Assessments: items,
				Total:       int64(totalCount),
			}
		},
	)

	return &AssessmentController{
		GenericController: genericController,
		svc:               service,
	}
}

func (ctl *AssessmentController) AssignRefLesson(c *gin.Context) {
	var req requests.AssessmentRefLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.data_invalid")
		return
	}

	if err := ctl.svc.AssignRefLesson(c, &req); err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, gin.H{
		"assessment_id": req.AssessmentID,
		"lesson_id":     req.LessonID,
		"course_id":     req.CourseID,
	}, nil, "")
}

func (ctl *AssessmentController) CreateWithCriteria(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.CreateAssessmentWithCriteriaRequest](c, func() *prot.CreateAssessmentWithCriteriaRequest {
		return &prot.CreateAssessmentWithCriteriaRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	assessment, err := ctl.svc.CreateWithCriteria(c, req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, assessment, nil, "")
}

func (ctl *AssessmentController) SaveScoreBulk(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveAssessmentScoreBulkRequest](c, func() *prot.SaveAssessmentScoreBulkRequest {
		return &prot.SaveAssessmentScoreBulkRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	if err := ctl.svc.SaveScoreBulk(c, req); err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, gin.H{"message": "Lưu điểm thành công"}, nil, "")
}

func (ctl *AssessmentController) GetPublish(c *gin.Context) {
	result, err := ctl.svc.GetPublish(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, result, nil, "")
}

func (ctl *AssessmentController) UpdatePublish(c *gin.Context) {
	result, err := ctl.svc.UpdatePublish(c)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, result, nil, "")
}
