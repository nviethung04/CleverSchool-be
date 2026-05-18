package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentAssessmentController struct {
	service services.DashboardStudentAssessmentService
}

func NewDashboardStudentAssessmentController(service services.DashboardStudentAssessmentService) *DashboardStudentAssessmentController {
	return &DashboardStudentAssessmentController{service: service}
}

func (ctl *DashboardStudentAssessmentController) GetStudentAssessments(c *gin.Context) {
	var req requests.DashboardStudentAssessmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	result, err := ctl.service.GetStudentAssessments(c, &req)
	utils.Respond(c, result, err, "")
}

