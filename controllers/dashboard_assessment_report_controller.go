package controllers

import (
	"be-cleverschool/requests"
	"be-cleverschool/services"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

// DashboardAssessmentReportController dùng riêng cho API /api/dashboard/assessment/report
type DashboardAssessmentReportController struct {
	svc services.DashboardAssessmentReportService
}

func NewDashboardAssessmentReportController() *DashboardAssessmentReportController {
	return &DashboardAssessmentReportController{
		svc: services.NewDashboardAssessmentReportService(),
	}
}

// GetAssessmentReport
// GET /api/dashboard/assessment/report?class_id=&class_main_id=&assessment_id=&page=&limit=
// Nếu class_main_id > 0 thì bỏ qua class_id và lấy học sinh từ tất cả classes có class_main_id đó
func (ctl *DashboardAssessmentReportController) GetAssessmentReport(c *gin.Context) {
	var req requests.AssessmentReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetAssessmentReport(c, &req)
	utils.Respond(c, resp, err, "")
}



