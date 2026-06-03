package controllers

import (
	"be-lms/database/db"
	"be-lms/jobs"
	"be-lms/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type InternalCommandController struct{}

func NewInternalCommandController() *InternalCommandController {
	return &InternalCommandController{}
}

// GenerateSchoolStatisticsRequest request body cho API
type GenerateSchoolStatisticsRequest struct {
	StartDate string `json:"start_date" binding:"required"` // Format: 2006-01-02
	EndDate   string `json:"end_date" binding:"required"`   // Format: 2006-01-02
}

// GenerateSchoolStatisticsResponse response cho API
type GenerateSchoolStatisticsResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	ProcessedAt  string `json:"processed_at"`
	Duration     string `json:"duration"`
	SchoolsCount int    `json:"schools_count,omitempty"`
}

// GenerateSchoolStatistics chạy cronjob tạo thống kê theo trường
func (c *InternalCommandController) GenerateSchoolStatistics(ctx *gin.Context) {
	var req GenerateSchoolStatisticsRequest

	// Parse request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, err, "Dữ liệu request không hợp lệ")
		return
	}

	// Parse ngày tháng
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày bắt đầu không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày kết thúc không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	// Validate thời gian
	if startDate.After(endDate) {
		utils.Respond(ctx, nil, nil, "Ngày bắt đầu không thể sau ngày kết thúc")
		return
	}

	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch cronjob
	cronJob := jobs.NewSchoolStatisticsCronJob()
	cronJob.RunForDateRange(startDate, endDate)

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tạo response
	response := GenerateSchoolStatisticsResponse{
		Success:     true,
		Message:     "Tạo thống kê theo trường thành công",
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}

// GenerateSchoolStatisticsLastWeek chạy cronjob cho tuần trước
func (c *InternalCommandController) GenerateSchoolStatisticsLastWeek(ctx *gin.Context) {
	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch cronjob cho tuần trước
	cronJob := jobs.NewSchoolStatisticsCronJob()
	cronJob.Run()

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tính toán thời gian tuần trước để hiển thị
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // Chủ nhật
		weekday = 7
	}

	mondayOfCurrentWeek := now.AddDate(0, 0, -weekday+1)
	mondayOfCurrentWeek = time.Date(mondayOfCurrentWeek.Year(), mondayOfCurrentWeek.Month(), mondayOfCurrentWeek.Day(), 0, 0, 0, 0, mondayOfCurrentWeek.Location())

	mondayOfLastWeek := mondayOfCurrentWeek.AddDate(0, 0, -7)
	sundayOfLastWeek := mondayOfLastWeek.AddDate(0, 0, 6)

	// Tạo response
	response := GenerateSchoolStatisticsResponse{
		Success:     true,
		Message:     "Tạo thống kê theo trường cho tuần trước thành công",
		StartDate:   mondayOfLastWeek.Format("2006-01-02"),
		EndDate:     sundayOfLastWeek.Format("2006-01-02"),
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}

// GetSchoolStatistics lấy thống kê đã tạo
func (c *InternalCommandController) GetSchoolStatistics(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	schoolIDStr := ctx.Query("school_id")

	// Validate tham số
	if startDateStr == "" || endDateStr == "" {
		utils.Respond(ctx, nil, nil, "Thiếu tham số start_date hoặc end_date")
		return
	}

	// Parse ngày tháng
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày bắt đầu không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày kết thúc không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	// Khởi tạo service
	cronJob := jobs.NewSchoolStatisticsCronJob()

	// Nếu có school_id, lấy thống kê của trường cụ thể
	if schoolIDStr != "" {
		schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
		if err != nil {
			utils.Respond(ctx, nil, err, "School ID không hợp lệ")
			return
		}

		// Lấy thống kê của trường cụ thể
		statistics, err := cronJob.GetSchoolStatistics(schoolID, startDate, endDate)
		if err != nil {
			utils.Respond(ctx, nil, err, "Không tìm thấy thống kê cho trường này")
			return
		}

		utils.Respond(ctx, statistics, nil, "")
		return
	}

	// Lấy thống kê của tất cả trường
	allStatistics, err := cronJob.GetAllSchoolStatistics(startDate, endDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Không tìm thấy thống kê")
		return
	}

	utils.Respond(ctx, allStatistics, nil, "")
}

// ====== COURSE STATISTICS API ======

type GenerateCourseStatisticsRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type GenerateCourseStatisticsResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	ProcessedAt string `json:"processed_at"`
	Duration    string `json:"duration"`
}

// GenerateCourseStatistics chạy cronjob để tạo thống kê theo khóa học cho khoảng thời gian cụ thể
func (c *InternalCommandController) GenerateCourseStatistics(ctx *gin.Context) {
	var req GenerateCourseStatisticsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, err, "Invalid request body")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày bắt đầu không hợp lệ (cần: YYYY-MM-DD)")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày kết thúc không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	if startDate.After(endDate) {
		utils.Respond(ctx, nil, nil, "Ngày bắt đầu không thể sau ngày kết thúc")
		return
	}

	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch cronjob
	cronJob := jobs.NewCourseStatisticsCronJob()
	cronJob.RunForDateRange(startDate, endDate)

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tạo response
	response := GenerateCourseStatisticsResponse{
		Success:     true,
		Message:     "Tạo thống kê theo khóa học thành công",
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}

// GenerateCourseStatisticsLastWeek chạy cronjob cho tuần trước
func (c *InternalCommandController) GenerateCourseStatisticsLastWeek(ctx *gin.Context) {
	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch cronjob cho tuần trước
	cronJob := jobs.NewCourseStatisticsCronJob()
	cronJob.Run()

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tính toán thời gian tuần trước để hiển thị
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // Chủ nhật
		weekday = 7
	}
	mondayOfCurrentWeek := now.AddDate(0, 0, -weekday+1)
	mondayOfCurrentWeek = time.Date(mondayOfCurrentWeek.Year(), mondayOfCurrentWeek.Month(), mondayOfCurrentWeek.Day(), 0, 0, 0, 0, mondayOfCurrentWeek.Location())
	mondayOfLastWeek := mondayOfCurrentWeek.AddDate(0, 0, -7)
	sundayOfLastWeek := mondayOfLastWeek.AddDate(0, 0, 6)

	// Tạo response
	response := GenerateCourseStatisticsResponse{
		Success:     true,
		Message:     "Tạo thống kê theo khóa học cho tuần trước thành công",
		StartDate:   mondayOfLastWeek.Format("2006-01-02"),
		EndDate:     sundayOfLastWeek.Format("2006-01-02"),
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}

// GetCourseStatistics lấy thống kê đã tạo
func (c *InternalCommandController) GetCourseStatistics(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	courseIDStr := ctx.Query("course_id")

	if startDateStr == "" || endDateStr == "" {
		utils.Respond(ctx, nil, nil, "Cần cung cấp cả ngày bắt đầu và ngày kết thúc")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày bắt đầu không hợp lệ (cần: YYYY-MM-DD)")
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "Format ngày kết thúc không hợp lệ (cần: YYYY-MM-DD)")
		return
	}

	// Khởi tạo service
	cronJob := jobs.NewCourseStatisticsCronJob()

	// Nếu có course_id, lấy thống kê của khóa học cụ thể
	if courseIDStr != "" {
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			utils.Respond(ctx, nil, err, "Course ID không hợp lệ")
			return
		}

		// Lấy thống kê của khóa học cụ thể
		statistics, err := cronJob.GetCourseStatistics(courseID, startDate, endDate)
		if err != nil {
			utils.Respond(ctx, nil, err, "Không tìm thấy thống kê cho khóa học này")
			return
		}

		utils.Respond(ctx, statistics, nil, "")
		return
	}

	// Lấy thống kê của tất cả khóa học
	allStatistics, err := cronJob.GetAllCourseStatistics(startDate, endDate)
	if err != nil {
		utils.Respond(ctx, nil, err, "Không tìm thấy thống kê")
		return
	}

	utils.Respond(ctx, allStatistics, nil, "")
}

// ====== DAILY STATISTICS API ======

// DailyJobResponse response cho daily jobs
type DailyJobResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ProcessedAt string `json:"processed_at"`
	Duration    string `json:"duration"`
	JobType     string `json:"job_type"`
}

// RunDailySchoolStatistics chạy daily school statistics job thủ công
func (c *InternalCommandController) RunDailySchoolStatistics(ctx *gin.Context) {
	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch daily job
	job := jobs.NewDailySchoolStatisticsCronJob()
	if err := job.Run(); err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi chạy Daily School Statistics Job")
		return
	}

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tạo response
	response := DailyJobResponse{
		Success:     true,
		Message:     "Daily School Statistics Job chạy thành công",
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
		JobType:     "daily_school_statistics",
	}

	utils.Respond(ctx, response, nil, "")
}

// RunDailyCourseStatistics chạy daily course statistics job thủ công
func (c *InternalCommandController) RunDailyCourseStatistics(ctx *gin.Context) {
	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Dispatch daily job
	job := jobs.NewDailyCourseStatisticsCronJob()
	if err := job.Run(); err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi chạy Daily Course Statistics Job")
		return
	}

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tạo response
	response := DailyJobResponse{
		Success:     true,
		Message:     "Daily Course Statistics Job chạy thành công",
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
		JobType:     "daily_course_statistics",
	}

	utils.Respond(ctx, response, nil, "")
}

// RunAllDailyStatistics chạy cả 2 daily jobs
func (c *InternalCommandController) RunAllDailyStatistics(ctx *gin.Context) {
	// Bắt đầu đo thời gian
	startTime := time.Now()

	// Chạy Daily School Statistics Job
	schoolJob := jobs.NewDailySchoolStatisticsCronJob()
	if err := schoolJob.Run(); err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi chạy Daily School Statistics Job")
		return
	}

	// Chạy Daily Course Statistics Job
	courseJob := jobs.NewDailyCourseStatisticsCronJob()
	if err := courseJob.Run(); err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi chạy Daily Course Statistics Job")
		return
	}

	// Tính thời gian xử lý
	duration := time.Since(startTime)

	// Tạo response
	response := DailyJobResponse{
		Success:     true,
		Message:     "Tất cả Daily Statistics Jobs chạy thành công",
		ProcessedAt: time.Now().Format("2006-01-02 15:04:05"),
		Duration:    duration.String(),
		JobType:     "all_daily_statistics",
	}

	utils.Respond(ctx, response, nil, "")
}

// ListAvailableJobs liệt kê các jobs có sẵn
func (c *InternalCommandController) ListAvailableJobs(ctx *gin.Context) {
	jobs := []map[string]interface{}{
		{
			"name":        "school-statistics-generate",
			"description": "Tạo thống kê theo trường cho khoảng thời gian cụ thể",
			"method":      "POST",
			"endpoint":    "/api/internal/school-statistics/generate",
			"body": map[string]string{
				"start_date": "YYYY-MM-DD",
				"end_date":   "YYYY-MM-DD",
			},
		},
		{
			"name":        "school-statistics-last-week",
			"description": "Tạo thống kê theo trường cho tuần trước",
			"method":      "POST",
			"endpoint":    "/api/internal/school-statistics/generate-last-week",
			"body":        nil,
		},
		{
			"name":        "course-statistics-generate",
			"description": "Tạo thống kê theo khóa học cho khoảng thời gian cụ thể",
			"method":      "POST",
			"endpoint":    "/api/internal/course-statistics/generate",
			"body": map[string]string{
				"start_date": "YYYY-MM-DD",
				"end_date":   "YYYY-MM-DD",
			},
		},
		{
			"name":        "course-statistics-last-week",
			"description": "Tạo thống kê theo khóa học cho tuần trước",
			"method":      "POST",
			"endpoint":    "/api/internal/course-statistics/generate-last-week",
			"body":        nil,
		},
		{
			"name":        "daily-school-statistics",
			"description": "Chạy Daily School Statistics Job (4 khoảng thời gian)",
			"method":      "POST",
			"endpoint":    "/api/internal/daily/school-statistics",
			"body":        nil,
		},
		{
			"name":        "daily-course-statistics",
			"description": "Chạy Daily Course Statistics Job (4 khoảng thời gian)",
			"method":      "POST",
			"endpoint":    "/api/internal/daily/course-statistics",
			"body":        nil,
		},
		{
			"name":        "daily-all-statistics",
			"description": "Chạy tất cả Daily Statistics Jobs",
			"method":      "POST",
			"endpoint":    "/api/internal/daily/all-statistics",
			"body":        nil,
		},
		{
			"name":        "list-jobs",
			"description": "Liệt kê tất cả jobs có sẵn",
			"method":      "GET",
			"endpoint":    "/api/internal/jobs",
			"body":        nil,
		},
	}

	utils.Respond(ctx, map[string]interface{}{
		"total_jobs": len(jobs),
		"jobs":       jobs,
	}, nil, "")
}

// ====== CLEAR DATA API ======

// ClearDataResponse response cho clear data
type ClearDataResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RecordsCount int64  `json:"records_count"`
	ProcessedAt  string `json:"processed_at"`
	Duration     string `json:"duration"`
}

// ClearDashboardSchoolsData xóa tất cả dữ liệu trong bảng dashboard_report_schools
func (c *InternalCommandController) ClearDashboardSchoolsData(ctx *gin.Context) {
	startTime := time.Now()

	// Đếm số records
	var count int64
	db.MasterDB.Table("dashboard_report_schools").Count(&count)

	// Xóa tất cả data
	if err := db.MasterDB.Exec("DELETE FROM dashboard_report_schools").Error; err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi xóa dữ liệu dashboard_report_schools")
		return
	}

	duration := time.Since(startTime)

	response := ClearDataResponse{
		Success:      true,
		Message:      "Xóa dữ liệu dashboard_report_schools thành công",
		RecordsCount: count,
		ProcessedAt:  time.Now().Format("2006-01-02 15:04:05"),
		Duration:     duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}

// ClearDashboardCoursesData xóa tất cả dữ liệu trong bảng dashboard_report_courses
func (c *InternalCommandController) ClearDashboardCoursesData(ctx *gin.Context) {
	startTime := time.Now()

	// Đếm số records
	var count int64
	db.MasterDB.Table("dashboard_report_courses").Count(&count)

	// Xóa tất cả data
	if err := db.MasterDB.Exec("DELETE FROM dashboard_report_courses").Error; err != nil {
		utils.Respond(ctx, nil, err, "Lỗi khi xóa dữ liệu dashboard_report_courses")
		return
	}

	duration := time.Since(startTime)

	response := ClearDataResponse{
		Success:      true,
		Message:      "Xóa dữ liệu dashboard_report_courses thành công",
		RecordsCount: count,
		ProcessedAt:  time.Now().Format("2006-01-02 15:04:05"),
		Duration:     duration.String(),
	}

	utils.Respond(ctx, response, nil, "")
}
