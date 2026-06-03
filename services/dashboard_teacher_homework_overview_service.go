package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

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
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy homework cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	overview, err := s.repo.GetOverview(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkOverviewStatsResponse{
		Overview: resources.DashboardTeacherHomeworkOverviewStatsResource(*overview),
	}, nil
}