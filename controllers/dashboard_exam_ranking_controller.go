package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardExamRankingController struct {
	svc services.DashboardExamRankingService
}

func NewDashboardExamRankingController() *DashboardExamRankingController {
	return &DashboardExamRankingController{
		svc: services.NewDashboardExamRankingService(),
	}
}

func (ctl *DashboardExamRankingController) GetExamRanking(c *gin.Context) {
	var req requests.DashboardExamRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetExamRanking(c, &req)
	utils.Respond(c, resp, err, "")
}
