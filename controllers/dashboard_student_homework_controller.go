package controllers

import (
	"be-cleverschool/requests"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentHomeworkController struct {
    service services.DashboardStudentHomeworkService
}

func NewDashboardStudentHomeworkController(service services.DashboardStudentHomeworkService) *DashboardStudentHomeworkController {
    return &DashboardStudentHomeworkController{service: service}
}

func (ctl *DashboardStudentHomeworkController) GetStudentHomeworkStats(c *gin.Context) {
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
    stats, err := ctl.service.GetStudentHomeworkStats(userID, req.CourseID, req.StartDate, req.EndDate)
    utils.Respond(c, stats, err, "")
}

