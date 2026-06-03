package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkUnscoredController struct {
	svc services.DashboardTeacherHomeworkUnscoredService
}

func NewDashboardTeacherHomeworkUnscoredController() *DashboardTeacherHomeworkUnscoredController {
	return &DashboardTeacherHomeworkUnscoredController{
		svc: services.NewDashboardTeacherHomeworkUnscoredService(),
	}
}

func (ctl *DashboardTeacherHomeworkUnscoredController) DashboardTeacherHomeworkUnscored(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkUnscoredListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetUnscoredHomeworks(c, &req)
	utils.Respond(c, resp, err, "")
}
