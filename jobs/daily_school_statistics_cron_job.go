package jobs

import (
	"fmt"
	"time"

	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
)

type DailySchoolStatisticsCronJob struct {
	service services.DashboardReportSchoolsService
}

func NewDailySchoolStatisticsCronJob() *DailySchoolStatisticsCronJob {
	return &DailySchoolStatisticsCronJob{}
}

func (j *DailySchoolStatisticsCronJob) initService() error {
	if j.service != nil {
		return nil
	}

	// Load config
	cfg := config.LoadConfig()

	// Connect to database
	if err := db.ConnectPostgres(cfg); err != nil {
		return fmt.Errorf("lỗi kết nối database: %v", err)
	}

	// Initialize repository and service
	repository := repositories.NewDashboardReportSchoolsRepository()
	j.service = services.NewDashboardReportSchoolsService(repository)

	return nil
}

// Run chạy job hàng ngày với 4 khoảng thời gian khác nhau
func (j *DailySchoolStatisticsCronJob) Run() error {
	if err := j.initService(); err != nil {
		return err
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	
	fmt.Printf("🚀 Bắt đầu Daily School Statistics Job - %s\n", now.Format("2006-01-02 15:04:05"))

	// Định nghĩa các khoảng thời gian
	dateRanges := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
	}{
		{
			name:      "Từ 15/9/2025 đến hôm qua",
			startDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
			endDate:   yesterday,
		},
		{
			name:      "Từ 8/9/2025 đến hôm qua", 
			startDate: time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC),
			endDate:   yesterday,
		},
		{
			name:      "Từ đầu tuần hiện tại đến cuối tuần hiện tại",
			startDate: j.getStartOfWeek(now),
			endDate:   j.getEndOfWeek(now),
		},
		{
			name:      "Từ 15/9/2025 đến cuối tuần trước",
			startDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
			endDate:   j.getEndOfWeek(yesterday),
		},
	}

	// Chạy từng khoảng thời gian
	for i, dateRange := range dateRanges {
		fmt.Printf("\n📊 [%d/4] Xử lý: %s\n", i+1, dateRange.name)
		fmt.Printf("   📅 Từ: %s đến: %s\n", 
			dateRange.startDate.Format("2006-01-02"), 
			dateRange.endDate.Format("2006-01-02"))

		startTime := time.Now()
		
		if err := j.service.GenerateSchoolStatistics(dateRange.startDate, dateRange.endDate); err != nil {
			fmt.Printf("❌ Lỗi khi tạo thống kê cho %s: %v\n", dateRange.name, err)
			continue
		}

		duration := time.Since(startTime)
		fmt.Printf("✅ Hoàn thành %s trong %v\n", dateRange.name, duration)
	}

	fmt.Printf("\n🎉 Hoàn thành Daily School Statistics Job\n")
	return nil
}

// getStartOfWeek lấy ngày đầu tuần (Thứ 2)
func (j *DailySchoolStatisticsCronJob) getStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Chủ nhật
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday - 1))
}

// getEndOfWeek lấy ngày cuối tuần (Chủ nhật)
func (j *DailySchoolStatisticsCronJob) getEndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Chủ nhật
		return t
	}
	return t.AddDate(0, 0, 7-weekday)
}

