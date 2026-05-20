package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardContestController struct {
	service services.DashboardContestService
}

func NewDashboardContestController(service services.DashboardContestService) *DashboardContestController {
	return &DashboardContestController{service: service}
}

// GetContestStats handles GET /api/dashboard/contests/stats
func (dcc *DashboardContestController) GetContestStats(c *gin.Context) {
	userID := int64(utils.GetCurrentUserId(c))
	if userID == 0 {
		utils.Respond(c, nil, nil, "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := dcc.service.GetContestStats(c, userID)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}
	utils.Respond(c, stats, nil, "")
}

// GetContestList handles GET /api/dashboard/contests/list
func (dcc *DashboardContestController) GetContestList(c *gin.Context) {
	var req requests.DashboardContestListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request", http.StatusBadRequest)
		return
	}

	userID := int64(utils.GetCurrentUserId(c))
	if userID == 0 {
		utils.Respond(c, nil, nil, "messages.unauthorized", http.StatusUnauthorized)
		return
	}
	req.UserID = userID
	list, total, err := dcc.service.GetContestList(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}
	utils.Respond(c, &prot.DashboardContestListResponse{Contests: list, Total: total}, nil, "")
}

// GetContestRoundStats handles GET /api/dashboard/contest-rounds/:id/stats
func (dcc *DashboardContestController) GetContestRoundStats(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id", http.StatusBadRequest)
		return
	}

	userID := int64(utils.GetCurrentUserId(c))
	if userID == 0 {
		utils.Respond(c, nil, nil, "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := dcc.service.GetContestRoundStats(c, int64(0), userID)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}
	utils.Respond(c, stats, nil, "")
}

// GetContestRoundList handles GET /api/dashboard/contest-rounds/:id/list
func (dcc *DashboardContestController) GetContestRoundList(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id", http.StatusBadRequest)
		return
	}

	var req requests.DashboardContestRoundListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request", http.StatusBadRequest)
		return
	}

	userID := int64(utils.GetCurrentUserId(c))
	if userID == 0 {
		utils.Respond(c, nil, nil, "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	req.UserID = userID
	req.ContestID = int64(0)
	list, total, err := dcc.service.GetContestRoundList(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}
	utils.Respond(c, &prot.DashboardContestRoundListResponse{Rounds: list, Total: total}, nil, "")
}
