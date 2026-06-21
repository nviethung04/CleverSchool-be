package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExerciseListController struct {
	svc services.DashboardTeacherExerciseListService
}

func NewDashboardTeacherExerciseListController() *DashboardTeacherExerciseListController {
	return &DashboardTeacherExerciseListController{
		svc: services.NewDashboardTeacherExerciseListService(),
	}
}

func (ctl *DashboardTeacherExerciseListController) GetStudentExerciseList(c *gin.Context) {
	var req requests.DashboardTeacherExerciseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	data, err := ctl.svc.GetStudentExerciseList(&req)
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

type DashboardTeacherExerciseStudentController struct {
	svc services.DashboardTeacherExerciseStudentService
}

func NewDashboardTeacherExerciseStudentController() *DashboardTeacherExerciseStudentController {
	return &DashboardTeacherExerciseStudentController{
		svc: services.NewDashboardTeacherExerciseStudentService(),
	}
}

func (ctl *DashboardTeacherExerciseStudentController) GetStudentStats(c *gin.Context) {
	var req requests.DashboardTeacherExerciseStudentStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	data, err := ctl.svc.GetStudentStats(&req)
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
