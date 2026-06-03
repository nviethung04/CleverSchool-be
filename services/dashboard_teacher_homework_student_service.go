package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"github.com/gin-gonic/gin"
)

type DashboardTeacherHomeworkStudentService interface {
	GetStudents(c *gin.Context, req *requests.DashboardTeacherHomeworkStudentRequest) (*prot.DashboardTeacherHomeworkStudentListResponse, error)
	GetStudentStats(c *gin.Context, req *requests.DashboardTeacherHomeworkStudentStatsRequest) (*prot.DashboardTeacherHomeworkStudentStatsResponse, error)
	GetHomeworkOverview(c *gin.Context, req *requests.DashboardTeacherHomeworkOverviewRequest) (*prot.DashboardTeacherHomeworkOverviewResponse, error)
}

type dashboardTeacherHomeworkStudentService struct {
	repo repositories.DashboardTeacherHomeworkStudentRepository
}

func NewDashboardTeacherHomeworkStudentService() DashboardTeacherHomeworkStudentService {
	return &dashboardTeacherHomeworkStudentService{
		repo: repositories.NewDashboardTeacherHomeworkStudentRepository(),
	}
}

func (s *dashboardTeacherHomeworkStudentService) GetStudents(c *gin.Context, req *requests.DashboardTeacherHomeworkStudentRequest) (*prot.DashboardTeacherHomeworkStudentListResponse, error) {
	students, totalCount, err := s.repo.GetStudents(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkStudentListResponse{
		Students: resources.DashboardTeacherHomeworkStudentCollection(students),
		Total:    totalCount,
	}, nil
}

func (s *dashboardTeacherHomeworkStudentService) GetStudentStats(c *gin.Context, req *requests.DashboardTeacherHomeworkStudentStatsRequest) (*prot.DashboardTeacherHomeworkStudentStatsResponse, error) {
	students, totalCount, err := s.repo.GetStudentStats(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkStudentStatsResponse{
		Students: resources.DashboardTeacherHomeworkStudentStatsCollection(students),
		Total:    totalCount,
	}, nil
}

func (s *dashboardTeacherHomeworkStudentService) GetHomeworkOverview(c *gin.Context, req *requests.DashboardTeacherHomeworkOverviewRequest) (*prot.DashboardTeacherHomeworkOverviewResponse, error) {
	overview, err := s.repo.GetHomeworkOverview(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherHomeworkOverviewResponse{
		Overview: resources.DashboardTeacherHomeworkOverviewResource(*overview),
	}, nil
} 