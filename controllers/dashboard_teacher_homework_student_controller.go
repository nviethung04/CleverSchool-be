package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkStudentController struct {
	svc services.DashboardTeacherHomeworkStudentService
}

func NewDashboardTeacherHomeworkStudentController() *DashboardTeacherHomeworkStudentController {
	return &DashboardTeacherHomeworkStudentController{
		svc: services.NewDashboardTeacherHomeworkStudentService(),
	}
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkStudent(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Get homework_id from URL parameter
	homeworkIDStr := c.Param("homework_id")
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Set the homework_id from URL parameter
	req.HomeworkID = homeworkID

	resp, err := ctl.svc.GetStudents(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkStats(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetStudentStats(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkOverview(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetHomeworkOverview(c, &req)
	utils.Respond(c, resp, err, "")
}
