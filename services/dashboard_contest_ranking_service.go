package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"

	"be-cleverschool/models"

	"github.com/gin-gonic/gin"
)

type DashboardContestRankingService interface {
	GetContestRanking(c *gin.Context, req *requests.DashboardContestRankingRequest) (*prot.DashboardContestRankingResponse, error)
}

type dashboardContestRankingService struct {
	repo repositories.DashboardContestRankingRepository
}

func NewDashboardContestRankingService() DashboardContestRankingService {
	return &dashboardContestRankingService{
		repo: repositories.NewDashboardContestRankingRepository(),
	}
}

func (s *dashboardContestRankingService) GetContestRanking(c *gin.Context, req *requests.DashboardContestRankingRequest) (*prot.DashboardContestRankingResponse, error) {
	// Lấy thông tin user hiện tại từ token
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Chỉ tính ranking nếu là học sinh (role_id = 3)
	var currentUserID int64 = 0
	if roleID == models.StudentRoleId && userID > 0 {
		currentUserID = int64(userID)
	}

	rankings, totalCount, currentRanking, err := s.repo.GetContestRanking(req, currentUserID)
	if err != nil {
		return nil, err
	}

	// Lấy score chart - sử dụng userID = 0 để lấy tất cả dữ liệu
	scoreChart, err := s.repo.GetContestScoreChart(req)
	if err != nil {
		return nil, err
	}

	// Convert score chart to protobuf
	var protoScoreChart []*prot.ContestScoreDistribution
	for _, chart := range scoreChart {
		protoScoreChart = append(protoScoreChart, &prot.ContestScoreDistribution{
			ScoreRange:   chart.ScoreRange,
			StudentCount: chart.StudentCount,
		})
	}

	return &prot.DashboardContestRankingResponse{
		Students:       resources.DashboardContestRankingCollection(rankings),
		Total:          totalCount,
		ScoreChart:     protoScoreChart,
		CurrentRanking: currentRanking, // Thêm current_ranking vào response
	}, nil
}


