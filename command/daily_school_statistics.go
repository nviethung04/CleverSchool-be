package command

import (
	"be-lms/jobs"
	"fmt"
	"log"
)

func DailySchoolStatisticsCommand() {
	fmt.Println("🚀 Chạy Daily School Statistics Job...")
	
	job := jobs.NewDailySchoolStatisticsCronJob()
	if err := job.Run(); err != nil {
		log.Fatalf("❌ Lỗi khi chạy Daily School Statistics Job: %v", err)
	}
	
	fmt.Println("✅ Hoàn thành Daily School Statistics Job")
}
