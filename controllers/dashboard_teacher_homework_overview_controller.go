package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkOverviewController struct {
	svc services.DashboardTeacherHomeworkOverviewService
}

func NewDashboardTeacherHomeworkOverviewController() *DashboardTeacherHomeworkOverviewController {
	return &DashboardTeacherHomeworkOverviewController{
		svc: services.NewDashboardTeacherHomeworkOverviewService(),
	}
}

func (ctl *DashboardTeacherHomeworkOverviewController) DashboardTeacherHomeworkOverview(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkOverviewStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetOverview(c, &req)
	utils.Respond(c, resp, err, "")
}