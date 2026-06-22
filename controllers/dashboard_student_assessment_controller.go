package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentAssessmentController struct {
	service services.DashboardStudentAssessmentService
}

func NewDashboardStudentAssessmentController(service services.DashboardStudentAssessmentService) *DashboardStudentAssessmentController {
	return &DashboardStudentAssessmentController{service: service}
}

func (ctl *DashboardStudentAssessmentController) GetStudentAssessmentList(c *gin.Context) {
	var req requests.DashboardStudentAssessmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	if req.CourseID == nil || *req.CourseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "course_id is required"})
		return
	}

	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}

	result, err := ctl.service.GetStudentAssessmentList(userID, *req.CourseID, req.SubjectID, req.Type, limit, page)
	utils.Respond(c, result, err, "")
}

func (ctl *DashboardStudentAssessmentController) GetStudentAssessments(c *gin.Context) {
	var req requests.DashboardStudentAssessmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	if req.CourseID == nil || *req.CourseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "course_id is required"})
		return
	}

	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}

	result, err := ctl.service.GetStudentAssessments(userID, *req.CourseID, req.SubjectID, req.Type, req.AssessmentID, limit, page)
	utils.Respond(c, result, err, "")
}
