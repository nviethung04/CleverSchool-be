package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkListController struct {
	svc services.DashboardTeacherHomeworkListService
}

func NewDashboardTeacherHomeworkListController() *DashboardTeacherHomeworkListController {
	return &DashboardTeacherHomeworkListController{
		svc: services.NewDashboardTeacherHomeworkListService(),
	}
}

func (ctl *DashboardTeacherHomeworkListController) GetHomeworkList(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetHomeworkList(c, &req)
	utils.Respond(c, resp, err, "")
}

