package command

import (
	"be-Clever School/jobs"
	"fmt"
	"log"
)

func DailyCourseStatisticsCommand() {
	fmt.Println("🚀 Chạy Daily Course Statistics Job...")
	
	job := jobs.NewDailyCourseStatisticsCronJob()
	if err := job.Run(); err != nil {
		log.Fatalf("❌ Lỗi khi chạy Daily Course Statistics Job: %v", err)
	}
	
	fmt.Println("✅ Hoàn thành Daily Course Statistics Job")
}
