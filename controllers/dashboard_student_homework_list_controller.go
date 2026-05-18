package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentHomeworkListController struct {
    service services.DashboardStudentHomeworkListService
}

func NewDashboardStudentHomeworkListController(service services.DashboardStudentHomeworkListService) *DashboardStudentHomeworkListController {
    return &DashboardStudentHomeworkListController{service: service}
}

func (ctl *DashboardStudentHomeworkListController) GetStudentHomeworkList(c *gin.Context) {
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
    stats, err := ctl.service.GetStudentHomeworkList(userID, req.CourseID, req.StartDate, req.EndDate, req.Limit, req.Page)
    utils.Respond(c, stats, err, "")
}
