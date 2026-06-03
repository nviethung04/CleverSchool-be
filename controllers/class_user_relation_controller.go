package controllers

import (
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type ClassUserRelationController struct {
	svc services.ClassUserRelationService
}

func NewClassUserRelationController() *ClassUserRelationController {
	return &ClassUserRelationController{
		svc: services.NewClassUserRelationService(),
	}
}

func (ctl *ClassUserRelationController) UpdateUserClassRelation(c *gin.Context) {
    req, err, message := utils.GetBody[*prot.ClassUserRelationRequest](c, func() *prot.ClassUserRelationRequest {
		return &prot.ClassUserRelationRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	// Lấy user_id từ GetCurrentUserId
	updatedBy := utils.GetCurrentUserId(c)

	resp, err := ctl.svc.UpdateUserClassRelation(c, req.ClassId, req.StudentIds, int64(updatedBy))
	utils.Respond(c, resp, err, "")
}
