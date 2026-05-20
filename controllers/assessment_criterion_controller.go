package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssessmentCriterionController struct {
	*GenericController[models.AssessmentCriterion, prot.AssessmentCriterion, *prot.AssessmentCriterionRequest]
	svc services.AssessmentCriterionService
}

func NewAssessmentCriterionController(service services.AssessmentCriterionService) *AssessmentCriterionController {
	resource := resources.NewAssessmentCriterionResource()
	adapter := NewAssessmentCriterionResourceAdapter(resource)

	genericController := NewGenericController(
		service,
		adapter,
		func() *prot.AssessmentCriterionRequest {
			return &prot.AssessmentCriterionRequest{}
		},
		func(items []*prot.AssessmentCriterion, totalCount uint64) interface{} {
			return &prot.AssessmentCriterionListResponse{
				Criteria: items,
				Total:    int64(totalCount),
			}
		},
	)

	return &AssessmentCriterionController{
		GenericController: genericController,
		svc:               service,
	}
}

func (ctl *AssessmentCriterionController) CreateBulk(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.CreateAssessmentCriteriaRequest](c, func() *prot.CreateAssessmentCriteriaRequest {
		return &prot.CreateAssessmentCriteriaRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	err = ctl.svc.CreateBulk(c, req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, gin.H{"message": "Tạo criteria thành công"}, nil, "")
}
