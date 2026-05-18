package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardAssessmentReportExcelController struct {
	service services.DashboardAssessmentReportExcelService
}

func NewDashboardAssessmentReportExcelController() *DashboardAssessmentReportExcelController {
	return &DashboardAssessmentReportExcelController{
		service: services.NewDashboardAssessmentReportExcelService(),
	}
}

// ReportExcel xuất file Excel cho assessment report
// POST /api/dashboard/assessment/report-excel
func (ctl *DashboardAssessmentReportExcelController) ReportExcel(c *gin.Context) {
	var req requests.DashboardAssessmentReportExcelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	if req.SchoolIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "school_ids is required",
		})
		return
	}

	if len(req.Subjects) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "subjects is required",
		})
		return
	}

	err := ctl.service.ExportExcel(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to export Excel",
			"error":   err.Error(),
		})
		return
	}
}

