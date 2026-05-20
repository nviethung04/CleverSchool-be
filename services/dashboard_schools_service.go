package services

import (
	"be-cleverschool/dto"
	"be-cleverschool/repositories"
)

type DashboardSchoolsService interface {
	GetDashboardSchools(selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error)
}

type dashboardSchoolsService struct {
	dashboardSchoolsRepo repositories.DashboardSchoolsRepository
}

func NewDashboardSchoolsService(dashboardSchoolsRepo repositories.DashboardSchoolsRepository) DashboardSchoolsService {
	return &dashboardSchoolsService{
		dashboardSchoolsRepo: dashboardSchoolsRepo,
	}
}

// GetDashboardSchools lấy danh sách dashboard schools với điều kiện cố định
func (s *dashboardSchoolsService) GetDashboardSchools(selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error) {
	return s.dashboardSchoolsRepo.GetDashboardSchools(selectedStartDate, selectedEndDate)
}

