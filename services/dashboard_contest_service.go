package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/requests"

	"github.com/gin-gonic/gin"
)

type DashboardContestService interface {
	GetContestStats(c *gin.Context, userID int64) (*prot.DashboardContestStatsResponse, error)
	GetContestList(c *gin.Context, req requests.DashboardContestListRequest) ([]*prot.DashboardContestItem, int64, error)
	GetContestRoundStats(c *gin.Context, contestRoundID int64, userID int64) (*prot.DashboardContestRoundStatsResponse, error)
	GetContestRoundList(c *gin.Context, req requests.DashboardContestRoundListRequest) ([]*prot.DashboardContestRoundItem, int64, error)
}

type dashboardContestService struct {
	// TODO: Add repositories when implemented
}

func NewDashboardContestService() DashboardContestService {
	return &dashboardContestService{}
}

func (s *dashboardContestService) GetContestStats(c *gin.Context, userID int64) (*prot.DashboardContestStatsResponse, error) {
	// TODO: Implement logic to get contest stats for a user
	return &prot.DashboardContestStatsResponse{}, nil
}

func (s *dashboardContestService) GetContestList(c *gin.Context, req requests.DashboardContestListRequest) ([]*prot.DashboardContestItem, int64, error) {
	// TODO: Implement logic to get contest list for a user
	return []*prot.DashboardContestItem{}, 0, nil
}

func (s *dashboardContestService) GetContestRoundStats(c *gin.Context, contestRoundID int64, userID int64) (*prot.DashboardContestRoundStatsResponse, error) {
	// TODO: Implement logic to get contest round stats for a user
	return &prot.DashboardContestRoundStatsResponse{}, nil
}

func (s *dashboardContestService) GetContestRoundList(c *gin.Context, req requests.DashboardContestRoundListRequest) ([]*prot.DashboardContestRoundItem, int64, error) {
	// TODO: Implement logic to get contest round list for a user
	return []*prot.DashboardContestRoundItem{}, 0, nil
}

