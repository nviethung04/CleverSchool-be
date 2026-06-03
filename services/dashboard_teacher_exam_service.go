package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardTeacherExamService interface {
	GetScoredExams(c *gin.Context, req *requests.DashboardTeacherExamScoredListRequest) (*prot.DashboardTeacherExamScoredListResponse, error)
}

type dashboardTeacherExamService struct {
	repo repositories.DashboardTeacherExamRepository
}

func NewDashboardTeacherExamService() DashboardTeacherExamService {
	return &dashboardTeacherExamService{
		repo: repositories.NewDashboardTeacherExamRepository(),
	}
}

func (s *dashboardTeacherExamService) GetScoredExams(c *gin.Context, req *requests.DashboardTeacherExamScoredListRequest) (*prot.DashboardTeacherExamScoredListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	// Nếu role_id = 2 hoặc 3, chỉ lấy exam cùng khóa
	onlyUserCourses := false
	if (roleID == 2 || roleID == 3) && userID > 0 {
		onlyUserCourses = true
	}

	exams, totalCount, err := s.repo.GetScoredExams(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherExamScoredListResponse{
		Exams: resources.DashboardTeacherExamScoredCollection(exams),
		Total: totalCount,
	}, nil
} 