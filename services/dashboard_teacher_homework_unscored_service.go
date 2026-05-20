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
	if req.StartDate <= 0 || req.EndDate <= 0 {
		return nil, errors.New("start_date and end_date are required")
	}

	userID := utils.GetCurrentUserId(c)
	onlyUserCourses := userID > 0

	homeworks, totalCount, err := s.repo.GetUnscoredHomeworks(int64(userID), onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkUnscoredListResponse{
		Homeworks: resources.DashboardTeacherHomeworkUnscoredCollection(homeworks),
		Total:     totalCount,
	}, nil
}

