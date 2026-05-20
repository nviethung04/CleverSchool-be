package command

import (
	"be-cleverschool/jobs"
	"flag"
	"fmt"
	"log"
	"time"
)

// GenerateCourseStatisticsCommand tạo thống kê theo khóa học cho khoảng thời gian được chỉ định
func GenerateCourseStatisticsCommand() {
	var startDateStr, endDateStr string
	var help bool

	flag.StringVar(&startDateStr, "start", "", "Ngày bắt đầu (format: 2006-01-02)")
	flag.StringVar(&endDateStr, "end", "", "Ngày kết thúc (format: 2006-01-02)")
	flag.BoolVar(&help, "help", false, "Hiển thị hướng dẫn sử dụng")
	flag.Parse()

	if help {
		fmt.Println("Sử dụng: go run main.go generate-course-statistics -start=<YYYY-MM-DD> -end=<YYYY-MM-DD>")
		return
	}

	if startDateStr == "" || endDateStr == "" {
		log.Fatal("Cần cung cấp cả ngày bắt đầu và ngày kết thúc. Dùng -help để xem hướng dẫn.")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		log.Fatalf("Format ngày bắt đầu không hợp lệ: %v", err)
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.Fatalf("Format ngày kết thúc không hợp lệ: %v", err)
	}

	// Validate thời gian
	if startDate.After(endDate) {
		log.Fatal("Start date không thể sau end date")
	}

	// Dispatch cronjob
	fmt.Printf("Bắt đầu tính toán thống kê khóa học từ %s đến %s...\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	startTime := time.Now()
	cronJob := jobs.NewCourseStatisticsCronJob()
	cronJob.RunForDateRange(startDate, endDate)

	duration := time.Since(startTime)
	fmt.Printf("Hoàn thành tính toán thống kê trong %v\n", duration)

	// Hiển thị kết quả
	statistics, err := cronJob.GetAllCourseStatistics(startDate, endDate)
	if err != nil {
		log.Printf("Lỗi khi lấy kết quả: %v", err)
		return
	}

	fmt.Printf("\nKết quả thống kê cho %d khóa học:\n", len(statistics))
	fmt.Println("ID Khóa | Tổng HS | Tổng GV | HS Active | GV Active | HS Hoàn thành BT | Teacher IDs")
	fmt.Println("--------|---------|---------|-----------|-----------|------------------|------------")
	for _, stat := range statistics {
		fmt.Printf("%-7d | %-7d | %-7d | %-9d | %-9d | %-16d | %s\n",
			stat.CourseID, stat.TotalStudents, stat.TotalTeachers,
			stat.ActiveStudents, stat.ActiveTeachers, stat.StudentsCompletedHomework, stat.TeacherIDs)
	}
}

