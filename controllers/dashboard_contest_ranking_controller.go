package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardContestRankingController struct {
	svc services.DashboardContestRankingService
}

func NewDashboardContestRankingController() *DashboardContestRankingController {
	return &DashboardContestRankingController{
		svc: services.NewDashboardContestRankingService(),
	}
}

func (ctl *DashboardContestRankingController) GetContestRanking(c *gin.Context) {
	var req requests.DashboardContestRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetContestRanking(c, &req)
	utils.Respond(c, resp, err, "")
}

