package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

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

func (ctl *DashboardTeacherHomeworkListController) GetStudentHomeworkList(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	data, err := ctl.svc.GetStudentHomeworkList(&req)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
		"error":   "",
	})
}
