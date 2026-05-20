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
	if req.StartDate <= 0 || req.EndDate <= 0 {
		return nil, errors.New("start_date and end_date are required")
	}

	userID := utils.GetCurrentUserId(c)
	onlyUserCourses := userID > 0

	homeworks, totalCount, err := s.repo.GetScoredHomeworks(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkScoredListResponse{
		Homeworks: resources.DashboardTeacherHomeworkScoredCollection(homeworks),
		Total:     totalCount,
	}, nil
}

