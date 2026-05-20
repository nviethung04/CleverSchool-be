package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkScoredController struct {
	svc services.DashboardTeacherHomeworkScoredService
}

func NewDashboardTeacherHomeworkScoredController() *DashboardTeacherHomeworkScoredController {
	return &DashboardTeacherHomeworkScoredController{
		svc: services.NewDashboardTeacherHomeworkScoredService(),
	}
}

func (ctl *DashboardTeacherHomeworkScoredController) DashboardTeacherHomeworkScored(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkScoredListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetScoredHomeworks(c, &req)
	utils.Respond(c, resp, err, "")
}
