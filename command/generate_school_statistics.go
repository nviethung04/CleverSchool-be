package command

import (
	"be-cleverschool/jobs"
	"flag"
	"fmt"
	"log"
	"time"
)

// GenerateSchoolStatisticsCommand tạo thống kê theo trường cho khoảng thời gian được chỉ định
func GenerateSchoolStatisticsCommand() {
	var startDateStr, endDateStr string
	var help bool

	flag.StringVar(&startDateStr, "start", "", "Ngày bắt đầu (format: 2006-01-02)")
	flag.StringVar(&endDateStr, "end", "", "Ngày kết thúc (format: 2006-01-02)")
	flag.BoolVar(&help, "help", false, "Hiển thị hướng dẫn sử dụng")
	flag.Parse()

	if help {
		fmt.Println("Sử dụng:")
		fmt.Println("  go run main.go generate-school-statistics -start=2025-01-01 -end=2025-01-07")
		fmt.Println("")
		fmt.Println("Tham số:")
		fmt.Println("  -start: Ngày bắt đầu (format: 2006-01-02)")
		fmt.Println("  -end: Ngày kết thúc (format: 2006-01-02)")
		fmt.Println("  -help: Hiển thị hướng dẫn này")
		return
	}

	// Validate tham số
	if startDateStr == "" || endDateStr == "" {
		log.Fatal("Vui lòng cung cấp cả start và end date")
	}

	// Parse ngày tháng
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		log.Fatalf("Lỗi parse start date: %v", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.Fatalf("Lỗi parse end date: %v", err)
	}

	// Validate thời gian
	if startDate.After(endDate) {
		log.Fatal("Start date không thể sau end date")
	}

	// Dispatch cronjob
	fmt.Printf("Bắt đầu tính toán thống kê từ %s đến %s...\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	
	startTime := time.Now()
	cronJob := jobs.NewSchoolStatisticsCronJob()
	cronJob.RunForDateRange(startDate, endDate)
	
	duration := time.Since(startTime)
	fmt.Printf("Hoàn thành tính toán thống kê trong %v\n", duration)

	// Hiển thị kết quả
	statistics, err := cronJob.GetAllSchoolStatistics(startDate, endDate)
	if err != nil {
		log.Printf("Lỗi khi lấy kết quả: %v", err)
		return
	}

	fmt.Printf("\nKết quả thống kê cho %d trường:\n", len(statistics))
	fmt.Println("ID Trường | Tổng HS | Tổng GV | HS Active | GV Active | HS Hoàn thành BT")
	fmt.Println("----------|---------|---------|-----------|-----------|------------------")
	
	for _, stat := range statistics {
		fmt.Printf("%-9d | %-7d | %-7d | %-9d | %-9d | %-16d\n",
			stat.SchoolID,
			stat.TotalStudents,
			stat.TotalTeachers,
			stat.ActiveStudents,
			stat.ActiveTeachers,
			stat.StudentsCompletedHomework)
	}
}

