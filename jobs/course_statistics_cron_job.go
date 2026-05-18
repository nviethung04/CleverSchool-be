package jobs

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"log"
	"time"
)

// CourseStatisticsCronJob chạy cronjob để tạo thống kê theo khóa học
type CourseStatisticsCronJob struct {
	service services.DashboardReportCoursesService
}

// GetCourseStatistics lấy thống kê của một khóa học cụ thể
func (j *CourseStatisticsCronJob) GetCourseStatistics(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error) {
	j.initService()
	return j.service.GetCourseStatistics(courseID, startDate, endDate)
}

// GetAllCourseStatistics lấy thống kê của tất cả khóa học
func (j *CourseStatisticsCronJob) GetAllCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error) {
	j.initService()
	return j.service.GetAllCourseStatistics(startDate, endDate)
}

// NewCourseStatisticsCronJob tạo instance mới của cronjob
func NewCourseStatisticsCronJob() *CourseStatisticsCronJob {
	return &CourseStatisticsCronJob{}
}

// initService khởi tạo service khi cần
func (j *CourseStatisticsCronJob) initService() {
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
		repo := repositories.NewDashboardReportCoursesRepository()
		j.service = services.NewDashboardReportCoursesService(repo)
	}
}

// Run chạy cronjob để tạo thống kê cho tuần trước
func (j *CourseStatisticsCronJob) Run() {
	log.Println("🔄 Bắt đầu chạy Course Statistics CronJob...")

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
	err := j.service.GenerateCourseStatistics(mondayOfLastWeek, sundayOfLastWeek)
	if err != nil {
		log.Printf("❌ Lỗi khi tạo thống kê: %v", err)
		return
	}

	log.Println("✅ Hoàn thành Course Statistics CronJob")
}

// RunForDateRange chạy cronjob cho khoảng thời gian cụ thể
func (j *CourseStatisticsCronJob) RunForDateRange(startDate, endDate time.Time) {
	startTime := time.Now()
	log.Printf("🔄 Bắt đầu chạy Course Statistics CronJob cho khoảng thời gian từ %s đến %s...",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Khởi tạo service
	j.initService()

	// Tạo thống kê
	err := j.service.GenerateCourseStatistics(startDate, endDate)
	if err != nil {
		log.Printf("❌ Lỗi khi tạo thống kê: %v", err)
		return
	}

	duration := time.Since(startTime)
	log.Printf("✅ Hoàn thành Course Statistics CronJob cho khoảng thời gian được chỉ định trong %v", duration)
}
