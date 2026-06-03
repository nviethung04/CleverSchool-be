package jobs

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/services"
	"log"
	"time"
)

// SchoolStatisticsCronJob chạy cronjob để tạo thống kê theo trường
type SchoolStatisticsCronJob struct {
	service services.DashboardReportSchoolsService
}

// GetSchoolStatistics lấy thống kê của một trường cụ thể
func (j *SchoolStatisticsCronJob) GetSchoolStatistics(schoolID int64, startDate, endDate time.Time) (*models.DashboardReportSchoolWeeks, error) {
	j.initService()
	return j.service.GetSchoolStatistics(schoolID, startDate, endDate)
}

// GetAllSchoolStatistics lấy thống kê của tất cả trường
func (j *SchoolStatisticsCronJob) GetAllSchoolStatistics(startDate, endDate time.Time) ([]models.DashboardReportSchoolWeeks, error) {
	j.initService()
	return j.service.GetAllSchoolStatistics(startDate, endDate)
}

// NewSchoolStatisticsCronJob tạo instance mới của cronjob
func NewSchoolStatisticsCronJob() *SchoolStatisticsCronJob {
	return &SchoolStatisticsCronJob{}
}

// initService khởi tạo service khi cần
func (j *SchoolStatisticsCronJob) initService() {
	if j.service == nil {
		// Khởi tạo database connection
		cfg := config.LoadConfig()
		config.InitDisks()
		config.InitLogger()
		
		// Connect to database
		if err := db.ConnectPostgres(cfg); err != nil {
			log.Fatalf("❌ Database connection failed: %v", err)
		}

		// Khởi tạo repository và service
		repo := repositories.NewDashboardReportSchoolsRepository()
		j.service = services.NewDashboardReportSchoolsService(repo)
	}
}

// Run chạy cronjob để tạo thống kê cho tuần trước
func (j *SchoolStatisticsCronJob) Run() {
	log.Println("🔄 Bắt đầu chạy School Statistics CronJob...")

	// Khởi tạo service
	j.initService()

	// Tính toán thời gian tuần trước
	now := time.Now()
	// Lấy thứ 2 của tuần hiện tại
	weekday := int(now.Weekday())
	if weekday == 0 { // Chủ nhật
		weekday = 7
	}
	
	// Thứ 2 của tuần hiện tại
	mondayOfCurrentWeek := now.AddDate(0, 0, -weekday+1)
	mondayOfCurrentWeek = time.Date(mondayOfCurrentWeek.Year(), mondayOfCurrentWeek.Month(), mondayOfCurrentWeek.Day(), 0, 0, 0, 0, mondayOfCurrentWeek.Location())
	
	// Thứ 2 của tuần trước
	mondayOfLastWeek := mondayOfCurrentWeek.AddDate(0, 0, -7)
	// Chủ nhật của tuần trước
	sundayOfLastWeek := mondayOfLastWeek.AddDate(0, 0, 6)

	log.Printf("📅 Tính toán thống kê cho tuần từ %s đến %s", 
		mondayOfLastWeek.Format("2006-01-02"), 
		sundayOfLastWeek.Format("2006-01-02"))

	// Tạo thống kê
	err := j.service.GenerateSchoolStatistics(mondayOfLastWeek, sundayOfLastWeek)
	if err != nil {
		log.Printf("❌ Lỗi khi tạo thống kê: %v", err)
		return
	}

	log.Println("✅ Hoàn thành School Statistics CronJob")
}

// RunForDateRange chạy cronjob cho khoảng thời gian cụ thể
func (j *SchoolStatisticsCronJob) RunForDateRange(startDate, endDate time.Time) {
	startTime := time.Now()
	log.Printf("🔄 Bắt đầu chạy School Statistics CronJob cho khoảng thời gian từ %s đến %s...",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Khởi tạo service
	j.initService()

	// Tạo thống kê
	err := j.service.GenerateSchoolStatistics(startDate, endDate)
	if err != nil {
		log.Printf("❌ Lỗi khi tạo thống kê: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ Hoàn thành School Statistics CronJob cho khoảng thời gian được chỉ định trong %v", duration)
}
