package controllers

import (
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardReportController struct {
	service services.DashboardReportService
}

func NewDashboardReportController() *DashboardReportController {
	repo := repositories.NewDashboardReportRepository()
	service := services.NewDashboardReportService(repo)
	return &DashboardReportController{service: service}
}

func (c *DashboardReportController) GetDashboardReport(ctx *gin.Context) {
	// Lấy query parameters
	activeStudentTimeStr := ctx.Query("active_student_time")
	activeTeacherTimeStr := ctx.Query("active_teacher_time")

	var activeStudentTime, activeTeacherTime *int64

	// Parse active_student_time
	if activeStudentTimeStr != "" {
		if time, err := strconv.ParseInt(activeStudentTimeStr, 10, 64); err == nil {
			activeStudentTime = &time
		}
	}

	// Parse active_teacher_time
	if activeTeacherTimeStr != "" {
		if time, err := strconv.ParseInt(activeTeacherTimeStr, 10, 64); err == nil {
			activeTeacherTime = &time
		}
	}

	report, err := c.service.GetDashboardReport(activeStudentTime, activeTeacherTime)
	if err != nil {
		utils.Respond(ctx, nil, err, "Failed to get dashboard report")
		return
	}

	// Convert protobuf response to JSON for utils.Respond
	utils.Respond(ctx, report, nil, "")
}

