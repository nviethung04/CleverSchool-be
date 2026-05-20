package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type TeachingPlanController struct {
	*GenericController[models.TeachingPlan, prot.TeachingPlan, *prot.TeachingPlanRequest]
	service services.TeachingPlanService
}

func NewTeachingPlanController(service services.TeachingPlanService) *TeachingPlanController {
	teachingPlanResource := resources.NewTeachingPlanResource()
	teachingPlanResourceAdapter := NewTeachingPlanResourceAdapter(teachingPlanResource)

	genericController := NewGenericController(
		service,
		teachingPlanResourceAdapter,
		func() *prot.TeachingPlanRequest {
			return &prot.TeachingPlanRequest{}
		},
		func(teachingPlans []*prot.TeachingPlan, totalCount uint64) interface{} {
			return &prot.TeachingPlansResponse{
				TeachingPlans: teachingPlans,
				TotalCount:    totalCount,
			}
		},
	)

	return &TeachingPlanController{
		GenericController: genericController,
		service:           service,
	}
}

func (tpc *TeachingPlanController) Approve(c *gin.Context) {
	req, _, _ := utils.GetBody[*prot.TeachingPlanApproveRequest](c, func() *prot.TeachingPlanApproveRequest {
		return &prot.TeachingPlanApproveRequest{}
	})

	teachingPlan, err := tpc.service.Approve(c, req)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	teachingPlanFormat := tpc.resource.FormatItem(teachingPlan)
	utils.Respond(c, teachingPlanFormat, err, "")
}

