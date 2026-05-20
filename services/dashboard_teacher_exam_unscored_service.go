package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamUnscoredService interface {
	GetUnscoredExams(c *gin.Context, req *requests.DashboardTeacherExamUnscoredListRequest) (*prot.DashboardTeacherExamUnscoredListResponse, error)
}

type dashboardTeacherExamUnscoredService struct {
	repo repositories.DashboardTeacherExamUnscoredRepository
}

func NewDashboardTeacherExamUnscoredService() DashboardTeacherExamUnscoredService {
	return &dashboardTeacherExamUnscoredService{
		repo: repositories.NewDashboardTeacherExamUnscoredRepository(),
	}
}

func (s *dashboardTeacherExamUnscoredService) GetUnscoredExams(c *gin.Context, req *requests.DashboardTeacherExamUnscoredListRequest) (*prot.DashboardTeacherExamUnscoredListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy exam cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	exams, totalCount, err := s.repo.GetUnscoredExams(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherExamUnscoredListResponse{
		Exams: resources.DashboardTeacherExamUnscoredCollection(exams),
		Total: totalCount,
	}, nil
} 
