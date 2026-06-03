package services

import (
	"be-lms/dto"
	"be-lms/repositories"
)

type DashboardCoursesService interface {
	GetDashboardCourses(schoolID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error)
}

type dashboardCoursesService struct {
	dashboardCoursesRepo repositories.DashboardCoursesRepository
}

func NewDashboardCoursesService(dashboardCoursesRepo repositories.DashboardCoursesRepository) DashboardCoursesService {
	return &dashboardCoursesService{
		dashboardCoursesRepo: dashboardCoursesRepo,
	}
}

// GetDashboardCourses lấy danh sách dashboard courses với điều kiện cố định
func (s *dashboardCoursesService) GetDashboardCourses(schoolID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error) {
	return s.dashboardCoursesRepo.GetDashboardCourses(schoolID, selectedStartDate, selectedEndDate)
}
