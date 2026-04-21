package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardStudentExamListController struct {
    service services.DashboardStudentExamListService
}

func NewDashboardStudentExamListController(service services.DashboardStudentExamListService) *DashboardStudentExamListController {
    return &DashboardStudentExamListController{service: service}
}

func (ctl *DashboardStudentExamListController) GetStudentExamList(c *gin.Context) {
    var req requests.DashboardStudentExamListRequest
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
    stats, err := ctl.service.GetStudentExamList(userID, req.CourseID, req.StartDate, req.EndDate, req.Limit, req.Page)
    utils.Respond(c, stats, err, "")
}
