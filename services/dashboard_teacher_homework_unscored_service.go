package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkUnscoredService interface {
	GetUnscoredHomeworks(c *gin.Context, req *requests.DashboardTeacherHomeworkUnscoredListRequest) (*prot.DashboardTeacherHomeworkUnscoredListResponse, error)
}

type dashboardTeacherHomeworkUnscoredService struct {
	repo repositories.DashboardTeacherHomeworkUnscoredRepository
}

func NewDashboardTeacherHomeworkUnscoredService() DashboardTeacherHomeworkUnscoredService {
	return &dashboardTeacherHomeworkUnscoredService{
		repo: repositories.NewDashboardTeacherHomeworkUnscoredRepository(),
	}
}

func (s *dashboardTeacherHomeworkUnscoredService) GetUnscoredHomeworks(c *gin.Context, req *requests.DashboardTeacherHomeworkUnscoredListRequest) (*prot.DashboardTeacherHomeworkUnscoredListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy homework cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	homeworks, totalCount, err := s.repo.GetUnscoredHomeworks(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkUnscoredListResponse{
		Homeworks: resources.DashboardTeacherHomeworkUnscoredCollection(homeworks),
		Total:     totalCount,
	}, nil
}
