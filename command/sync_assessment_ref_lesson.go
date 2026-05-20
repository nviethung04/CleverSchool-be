package command

import (
	"be-cleverschool/services"
	"fmt"
	"log"
)

// SyncAssessmentRefLessonCommand xử lý đồng bộ assessment_ref_lessons
// Gọi service với programID = 0 để xử lý tất cả
func SyncAssessmentRefLessonCommand() {
	fmt.Println("🔄 Bắt đầu đồng bộ assessment_ref_lessons...")

	syncService := services.NewAssessmentRefLessonSyncService()
	totalProcessed, totalInserted, err := syncService.SyncAssessmentRefLessons(0)

	if err != nil {
		log.Printf("❌ Lỗi đồng bộ assessment_ref_lessons: %v", err)
		return
	}

	fmt.Printf("\n✅ Hoàn thành!\n")
	fmt.Printf("📊 Tổng kết:\n")
	fmt.Printf("  - Số cặp đã xử lý: %d\n", totalProcessed)
	fmt.Printf("  - Số records đã thêm: %d\n", totalInserted)
}


