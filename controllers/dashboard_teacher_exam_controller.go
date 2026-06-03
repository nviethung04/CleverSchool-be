package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamController struct {
	svc services.DashboardTeacherExamService
}

func NewDashboardTeacherExamController() *DashboardTeacherExamController {
	return &DashboardTeacherExamController{
		svc: services.NewDashboardTeacherExamService(),
	}
}

func (ctl *DashboardTeacherExamController) DashboardTeacherExamScored(c *gin.Context) {
	var req requests.DashboardTeacherExamScoredListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetScoredExams(c, &req)
	utils.Respond(c, resp, err, "")
}
