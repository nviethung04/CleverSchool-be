package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamOverviewService interface {
	GetOverview(c *gin.Context, req *requests.DashboardTeacherExamOverviewRequest) (*prot.DashboardTeacherExamOverviewResponse, error)
}

type dashboardTeacherExamOverviewService struct {
	repo repositories.DashboardTeacherExamOverviewRepository
}

func NewDashboardTeacherExamOverviewService() DashboardTeacherExamOverviewService {
	return &dashboardTeacherExamOverviewService{
		repo: repositories.NewDashboardTeacherExamOverviewRepository(),
	}
}

func (s *dashboardTeacherExamOverviewService) GetOverview(c *gin.Context, req *requests.DashboardTeacherExamOverviewRequest) (*prot.DashboardTeacherExamOverviewResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy exam cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	overview, err := s.repo.GetOverview(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherExamOverviewResponse{
		Overview: resources.DashboardTeacherExamOverviewResource(*overview),
	}, nil
} 