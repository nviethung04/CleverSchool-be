package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
)

type DashboardReportService interface {
	GetDashboardReport(activeStudentTime, activeTeacherTime *int64) (*prot.DashboardReportResponse, error)
}

type dashboardReportService struct {
	repo repositories.DashboardReportRepository
}

func NewDashboardReportService(repo repositories.DashboardReportRepository) DashboardReportService {
	return &dashboardReportService{repo: repo}
}

func (s *dashboardReportService) GetDashboardReport(activeStudentTime, activeTeacherTime *int64) (*prot.DashboardReportResponse, error) {
	// Lấy dữ liệu từ repository (DTO)
	dtoData, err := s.repo.GetDashboardReport(activeStudentTime, activeTeacherTime)
	if err != nil {
		return nil, err
	}

	// Convert DTO sang protobuf
	protoData := &prot.DashboardReportData{
		TotalSchools:                dtoData.TotalSchools,
		TotalClasses:                dtoData.TotalClasses,
		TotalUsers:                  dtoData.TotalUsers,
		StudentActive:               dtoData.StudentActive,
		TeacherActive:               dtoData.TeacherActive,
		StudentActiveWeekly:         dtoData.StudentActiveWeekly,
		TeacherActiveWeekly:         dtoData.TeacherActiveWeekly,
		FrequentLoginStudents:       dtoData.FrequentLoginStudents,
		FrequentLoginStudentsWeekly: dtoData.FrequentLoginStudentsWeekly,
	}

	// Tạo response
	response := &prot.DashboardReportResponse{
		Data: protoData,
	}

	return response, nil
}

