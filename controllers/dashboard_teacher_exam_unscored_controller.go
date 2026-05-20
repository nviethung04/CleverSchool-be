package controllers

import (
	"be-cleverschool/requests"
	"be-cleverschool/services"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamUnscoredController struct {
	svc services.DashboardTeacherExamUnscoredService
}

func NewDashboardTeacherExamUnscoredController() *DashboardTeacherExamUnscoredController {
	return &DashboardTeacherExamUnscoredController{
		svc: services.NewDashboardTeacherExamUnscoredService(),
	}
}

func (ctl *DashboardTeacherExamUnscoredController) DashboardTeacherExamUnscored(c *gin.Context) {
	var req requests.DashboardTeacherExamUnscoredListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetUnscoredExams(c, &req)
	utils.Respond(c, resp, err, "")
}

