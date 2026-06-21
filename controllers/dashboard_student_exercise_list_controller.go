package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentExerciseListController struct {
	service services.DashboardStudentExerciseListService
}

func NewDashboardStudentExerciseListController(service services.DashboardStudentExerciseListService) *DashboardStudentExerciseListController {
	return &DashboardStudentExerciseListController{service: service}
}

func (ctl *DashboardStudentExerciseListController) GetStudentExerciseList(c *gin.Context) {
	var req requests.DashboardStudentHomeworkListRequest
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
	stats, err := ctl.service.GetStudentExerciseList(userID, req.CourseID, req.StartDate, req.EndDate, req.Limit, req.Page)
	utils.Respond(c, stats, err, "")
}
