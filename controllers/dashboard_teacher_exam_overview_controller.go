package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamOverviewController struct {
	svc services.DashboardTeacherExamOverviewService
}

func NewDashboardTeacherExamOverviewController() *DashboardTeacherExamOverviewController {
	return &DashboardTeacherExamOverviewController{
		svc: services.NewDashboardTeacherExamOverviewService(),
	}
}

func (ctl *DashboardTeacherExamOverviewController) DashboardTeacherExamOverview(c *gin.Context) {
	var req requests.DashboardTeacherExamOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetOverview(c, &req)
	utils.Respond(c, resp, err, "")
}
