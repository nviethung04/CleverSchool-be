package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentExamController struct {
    service services.DashboardStudentExamService
}

func NewDashboardStudentExamController(service services.DashboardStudentExamService) *DashboardStudentExamController {
    return &DashboardStudentExamController{service: service}
}

func (ctl *DashboardStudentExamController) GetStudentExamStats(c *gin.Context) {
    var req requests.DashboardStudentExamRequest
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
    stats, err := ctl.service.GetStudentExamStats(userID, req.CourseID, req.StartDate, req.EndDate)
    utils.Respond(c, stats, err, "")
}
