package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkListService interface {
	GetHomeworkList(c *gin.Context, req *requests.DashboardTeacherHomeworkListRequest) (*prot.DashboardTeacherHomeworkListResponse, error)
}

type dashboardTeacherHomeworkListService struct {
	repo repositories.DashboardTeacherHomeworkListRepository
}

func NewDashboardTeacherHomeworkListService() DashboardTeacherHomeworkListService {
	return &dashboardTeacherHomeworkListService{
		repo: repositories.NewDashboardTeacherHomeworkListRepository(),
	}
}

func (s *dashboardTeacherHomeworkListService) GetHomeworkList(c *gin.Context, req *requests.DashboardTeacherHomeworkListRequest) (*prot.DashboardTeacherHomeworkListResponse, error) {
	homeworks, totalCount, err := s.repo.GetHomeworkList(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkListResponse{
		Homeworks: resources.DashboardTeacherHomeworkListCollection(homeworks),
		Total:     totalCount,
	}, nil
}


