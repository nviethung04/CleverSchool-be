package services

import (
	"errors"

	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkOverviewService interface {
	GetOverview(c *gin.Context, req *requests.DashboardTeacherHomeworkOverviewStatsRequest) (*prot.DashboardTeacherHomeworkOverviewStatsResponse, error)
}

type dashboardTeacherHomeworkOverviewService struct {
	repo repositories.DashboardTeacherHomeworkOverviewRepository
}

func NewDashboardTeacherHomeworkOverviewService() DashboardTeacherHomeworkOverviewService {
	return &dashboardTeacherHomeworkOverviewService{
		repo: repositories.NewDashboardTeacherHomeworkOverviewRepository(),
	}
}

func (s *dashboardTeacherHomeworkOverviewService) GetOverview(c *gin.Context, req *requests.DashboardTeacherHomeworkOverviewStatsRequest) (*prot.DashboardTeacherHomeworkOverviewStatsResponse, error) {
	if req.StartDate <= 0 || req.EndDate <= 0 {
		return nil, errors.New("start_date and end_date are required")
	}

	userID := utils.GetCurrentUserId(c)
	onlyUserCourses := userID > 0

	overview, err := s.repo.GetOverview(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkOverviewStatsResponse{
		Overview: resources.DashboardTeacherHomeworkOverviewStatsResource(*overview),
	}, nil
}
