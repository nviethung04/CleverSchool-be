package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"fmt"
	"time"
)

type DashboardReportSchoolsService interface {
	GenerateSchoolStatistics(startDate, endDate time.Time) error
	GetSchoolStatistics(schoolID int64, startDate, endDate time.Time) (*models.DashboardReportSchoolWeeks, error)
	GetAllSchoolStatistics(startDate, endDate time.Time) ([]models.DashboardReportSchoolWeeks, error)
}

type dashboardReportSchoolsService struct {
	repo repositories.DashboardReportSchoolsRepository
}

func NewDashboardReportSchoolsService(repo repositories.DashboardReportSchoolsRepository) DashboardReportSchoolsService {
	return &dashboardReportSchoolsService{repo: repo}
}

func (s *dashboardReportSchoolsService) GenerateSchoolStatistics(startDate, endDate time.Time) error {
	// Validate thời gian
	if startDate.After(endDate) {
		return fmt.Errorf("start_date không thể sau end_date")
	}

	// Tính toán thống kê cho tất cả trường
	statistics, err := s.repo.CalculateSchoolStatistics(startDate, endDate)
	if err != nil {
		return fmt.Errorf("lỗi khi tính toán thống kê: %v", err)
	}

	// Lưu thống kê vào database
	if err := s.repo.SaveSchoolStatistics(statistics); err != nil {
		return fmt.Errorf("lỗi khi lưu thống kê: %v", err)
	}

	return nil
}

func (s *dashboardReportSchoolsService) GetSchoolStatistics(schoolID int64, startDate, endDate time.Time) (*models.DashboardReportSchoolWeeks, error) {
	return s.repo.GetExistingReport(schoolID, startDate, endDate)
}

func (s *dashboardReportSchoolsService) GetAllSchoolStatistics(startDate, endDate time.Time) ([]models.DashboardReportSchoolWeeks, error) {
	// Lấy tất cả báo cáo trong khoảng thời gian
	var statistics []models.DashboardReportSchoolWeeks
	err := db.ReplicaDB.Where("start_date = ? AND end_date = ?", startDate, endDate).
		Find(&statistics).Error
	if err != nil {
		return nil, err
	}
	return statistics, nil
}
