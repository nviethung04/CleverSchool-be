package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentExerciseController struct {
	service services.DashboardStudentExerciseService
}

func NewDashboardStudentExerciseController(service services.DashboardStudentExerciseService) *DashboardStudentExerciseController {
	return &DashboardStudentExerciseController{service: service}
}

func (ctl *DashboardStudentExerciseController) GetStudentExerciseStats(c *gin.Context) {
	var req requests.DashboardStudentHomeworkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}
	stats, err := ctl.service.GetStudentExerciseStats(userID, req.CourseID, req.StartDate, req.EndDate)
	utils.Respond(c, stats, err, "")
}
