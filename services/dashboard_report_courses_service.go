package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"fmt"
	"time"
)

type DashboardReportCoursesService interface {
	GenerateCourseStatistics(startDate, endDate time.Time) error
	GetCourseStatistics(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error)
	GetAllCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error)
}

type dashboardReportCoursesService struct {
	repo repositories.DashboardReportCoursesRepository
}

func NewDashboardReportCoursesService(repo repositories.DashboardReportCoursesRepository) DashboardReportCoursesService {
	return &dashboardReportCoursesService{repo: repo}
}

func (s *dashboardReportCoursesService) GenerateCourseStatistics(startDate, endDate time.Time) error {
	// Validate thời gian
	if startDate.After(endDate) {
		return fmt.Errorf("start_date không thể sau end_date")
	}

	// Tính toán thống kê
	statistics, err := s.repo.CalculateCourseStatistics(startDate, endDate)
	if err != nil {
		return fmt.Errorf("lỗi khi tính toán thống kê: %w", err)
	}

	// Lưu vào database
	if err := s.repo.SaveCourseStatistics(statistics); err != nil {
		return fmt.Errorf("lỗi khi lưu thống kê: %w", err)
	}

	return nil
}

func (s *dashboardReportCoursesService) GetCourseStatistics(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error) {
	return s.repo.GetExistingReport(courseID, startDate, endDate)
}

func (s *dashboardReportCoursesService) GetAllCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error) {
	// Lấy tất cả báo cáo trong khoảng thời gian
	var statistics []models.DashboardReportCourses
	err := db.ReplicaDB.Where("start_date = ? AND end_date = ?", startDate, endDate).
		Find(&statistics).Error
	if err != nil {
		return nil, err
	}
	return statistics, nil
}

