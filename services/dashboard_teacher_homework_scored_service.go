package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkScoredService interface {
	GetScoredHomeworks(c *gin.Context, req *requests.DashboardTeacherHomeworkScoredListRequest) (*prot.DashboardTeacherHomeworkScoredListResponse, error)
}

type dashboardTeacherHomeworkScoredService struct {
	repo repositories.DashboardTeacherHomeworkScoredRepository
}

func NewDashboardTeacherHomeworkScoredService() DashboardTeacherHomeworkScoredService {
	return &dashboardTeacherHomeworkScoredService{
		repo: repositories.NewDashboardTeacherHomeworkScoredRepository(),
	}
}

func (s *dashboardTeacherHomeworkScoredService) GetScoredHomeworks(c *gin.Context, req *requests.DashboardTeacherHomeworkScoredListRequest) (*prot.DashboardTeacherHomeworkScoredListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy homework cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	homeworks, totalCount, err := s.repo.GetScoredHomeworks(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkScoredListResponse{
		Homeworks: resources.DashboardTeacherHomeworkScoredCollection(homeworks),
		Total:     totalCount,
	}, nil
}
