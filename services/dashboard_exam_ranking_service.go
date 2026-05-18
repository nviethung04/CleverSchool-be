package services

import (
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"be-Clever School/models"

	"github.com/gin-gonic/gin"
)

type DashboardExamRankingService interface {
	GetExamRanking(c *gin.Context, req *requests.DashboardExamRankingRequest) (*prot.DashboardExamRankingResponse, error)
}

type dashboardExamRankingService struct {
	repo repositories.DashboardExamRankingRepository
}

func NewDashboardExamRankingService() DashboardExamRankingService {
	return &dashboardExamRankingService{
		repo: repositories.NewDashboardExamRankingRepository(),
	}
}

func (s *dashboardExamRankingService) GetExamRanking(c *gin.Context, req *requests.DashboardExamRankingRequest) (*prot.DashboardExamRankingResponse, error) {
	// Lấy thông tin user hiện tại từ token
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Chỉ tính ranking nếu là học sinh (role_id = 3)
	var currentUserID int64 = 0
	if roleID == models.StudentRoleId && userID > 0 {
		currentUserID = int64(userID)
	}

	rankings, totalCount, currentRanking, err := s.repo.GetExamRanking(req, currentUserID)
	if err != nil {
		return nil, err
	}

	// Lấy score chart - sử dụng userID = 0 để lấy tất cả dữ liệu
	scoreChart, err := s.repo.GetExamScoreChart(req)
	if err != nil {
		return nil, err
	}

	// Convert score chart to protobuf
	var protoScoreChart []*prot.ExamScoreDistribution
	for _, chart := range scoreChart {
		protoScoreChart = append(protoScoreChart, &prot.ExamScoreDistribution{
			ScoreRange:   chart.ScoreRange,
			StudentCount: chart.StudentCount,
		})
	}

	return &prot.DashboardExamRankingResponse{
		Students:       resources.DashboardExamRankingCollection(rankings),
		Total:          totalCount,
		ScoreChart:     protoScoreChart,
		CurrentRanking: currentRanking, // Thêm current_ranking vào response
	}, nil
}
