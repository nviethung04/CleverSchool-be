package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AssessmentCriteriaGroupController struct {
	svc services.AssessmentCriteriaGroupService
	*GenericController[models.AssessmentCriteriaGroup, prot.AssessmentCriteriaGroup, *prot.AssessmentCriteriaGroupRequest]
}

func NewAssessmentCriteriaGroupController(service services.AssessmentCriteriaGroupService) *AssessmentCriteriaGroupController {
	resource := resources.NewAssessmentCriteriaGroupResource()
	adapter := NewAssessmentCriteriaGroupResourceAdapter(resource)

	genericController := NewGenericController(
		service,
		adapter,
		func() *prot.AssessmentCriteriaGroupRequest {
			return &prot.AssessmentCriteriaGroupRequest{}
		},
		func(items []*prot.AssessmentCriteriaGroup, totalCount uint64) interface{} {
			return &prot.AssessmentCriteriaGroupListResponse{
				Groups: items,
				Total:  int64(totalCount),
			}
		},
	)

	return &AssessmentCriteriaGroupController{
		GenericController: genericController,
		svc:               service,
	}
}

func (ctl *AssessmentCriteriaGroupController) CreateWithCriteria(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.CreateAssessmentCriteriaGroupWithCriteriaRequest](c, func() *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest {
		return &prot.CreateAssessmentCriteriaGroupWithCriteriaRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	group, err := ctl.svc.CreateWithCriteria(c, req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, group, nil, "")
}

func (ctl *AssessmentCriteriaGroupController) UpdateWithCriteria(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id", http.StatusBadRequest)
		return
	}

	req, err, message := utils.GetBody[*prot.CreateAssessmentCriteriaGroupWithCriteriaRequest](c, func() *prot.CreateAssessmentCriteriaGroupWithCriteriaRequest {
		return &prot.CreateAssessmentCriteriaGroupWithCriteriaRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	group, err := ctl.svc.UpdateWithCriteria(c, id, req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(c, group, nil, "")
}

