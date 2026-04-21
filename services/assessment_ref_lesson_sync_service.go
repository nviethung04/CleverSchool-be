package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

type AssessmentRefLessonSyncService interface {
	SyncAssessmentRefLessons(programID int64) (totalProcessed int, totalInserted int, err error)
}

type assessmentRefLessonSyncService struct{}

func NewAssessmentRefLessonSyncService() AssessmentRefLessonSyncService {
	return &assessmentRefLessonSyncService{}
}

// SyncAssessmentRefLessons đồng bộ assessment_ref_lessons
// Với mỗi cặp (assessment_id, lesson_id) distinct, lấy các course_id cùng program với lesson đó
// và thêm vào bảng assessment_ref_lessons (nếu chưa tồn tại)
// Nếu programID = 0 thì xử lý tất cả, nếu programID > 0 thì chỉ xử lý các lesson thuộc program đó
func (s *assessmentRefLessonSyncService) SyncAssessmentRefLessons(programID int64) (int, int, error) {
	// Bước 1: Xóa các cặp (lesson_id, assessment_id) không có hàng nào có course_id = 0
	log.Printf("🧹 Bắt đầu xóa các cặp không có course_id = 0...")
	
	// Lấy tất cả distinct cặp (assessment_id, lesson_id) có course_id = 0
	var pairsWithCourseZero []struct {
		AssessmentID int64 `gorm:"column:assessment_id"`
		LessonID     int64 `gorm:"column:lesson_id"`
	}
	
	if err := db.MasterDB.Table("assessment_ref_lessons").
		Select("DISTINCT assessment_id, lesson_id").
		Where("course_id = 0 OR course_id IS NULL").
		Find(&pairsWithCourseZero).Error; err != nil {
		return 0, 0, fmt.Errorf("lỗi lấy các cặp có course_id = 0: %w", err)
	}
	
	// Tạo map các cặp có course_id = 0
	pairsWithZeroMap := make(map[string]bool)
	for _, pair := range pairsWithCourseZero {
		key := fmt.Sprintf("%d_%d", pair.AssessmentID, pair.LessonID)
		pairsWithZeroMap[key] = true
	}
	
	// Lấy tất cả distinct cặp (assessment_id, lesson_id)
	var allPairs []struct {
		AssessmentID int64 `gorm:"column:assessment_id"`
		LessonID     int64 `gorm:"column:lesson_id"`
	}
	
	allPairsQuery := db.MasterDB.Table("assessment_ref_lessons").
		Select("DISTINCT assessment_id, lesson_id")
	
	// Nếu có programID, filter các lesson thuộc program đó
	if programID > 0 {
		allPairsQuery = allPairsQuery.
			Joins("JOIN lessons l ON l.id = assessment_ref_lessons.lesson_id").
			Joins("JOIN chapters ch ON ch.id = l.chapter_id").
			Where("ch.program_id = ? AND l.deleted_at IS NULL AND ch.deleted_at IS NULL", programID)
	}
	
	if err := allPairsQuery.Find(&allPairs).Error; err != nil {
		return 0, 0, fmt.Errorf("lỗi lấy tất cả các cặp: %w", err)
	}
	
	// Tìm các cặp không có course_id = 0
	type pairT struct {
		AssessmentID int64
		LessonID     int64
	}
	var pairsToDelete []pairT
	for _, pair := range allPairs {
		key := fmt.Sprintf("%d_%d", pair.AssessmentID, pair.LessonID)
		if !pairsWithZeroMap[key] {
			pairsToDelete = append(pairsToDelete, pairT{AssessmentID: pair.AssessmentID, LessonID: pair.LessonID})
		}
	}
	
	// Xóa tất cả các hàng của các cặp này
	totalDeleted := 0
	if len(pairsToDelete) > 0 {
		for _, pair := range pairsToDelete {
			result := db.MasterDB.
				Where("assessment_id = ? AND lesson_id = ?", pair.AssessmentID, pair.LessonID).
				Delete(&models.AssessmentRefLesson{})
			
			if result.Error != nil {
				log.Printf("⚠️ Lỗi xóa records cho assessment_id %d, lesson_id %d: %v", pair.AssessmentID, pair.LessonID, result.Error)
				continue
			}
			
			totalDeleted += int(result.RowsAffected)
		}
		log.Printf("🗑️ Đã xóa %d hàng từ %d cặp không có course_id = 0", totalDeleted, len(pairsToDelete))
	} else {
		log.Printf("✅ Không có cặp nào cần xóa")
	}

	// Bước 2: Lấy distinct các cặp (assessment_id, lesson_id) từ assessment_ref_lessons
	type AssessmentLessonPair struct {
		AssessmentID int64 `gorm:"column:assessment_id"`
		LessonID     int64 `gorm:"column:lesson_id"`
	}

	var pairs []AssessmentLessonPair
	query := db.MasterDB.Table("assessment_ref_lessons").
		Select("DISTINCT assessment_ref_lessons.assessment_id, assessment_ref_lessons.lesson_id")

	// Nếu có programID, filter các lesson thuộc program đó
	if programID > 0 {
		query = query.
			Joins("JOIN lessons l ON l.id = assessment_ref_lessons.lesson_id").
			Joins("JOIN chapters ch ON ch.id = l.chapter_id").
			Where("ch.program_id = ? AND l.deleted_at IS NULL AND ch.deleted_at IS NULL", programID)
	}

	if err := query.Find(&pairs).Error; err != nil {
		return 0, 0, fmt.Errorf("lỗi lấy distinct assessment_id và lesson_id: %w", err)
	}

	log.Printf("📊 Tìm thấy %d cặp (assessment_id, lesson_id)", len(pairs))

	totalProcessed := 0
	totalInserted := 0

	// Xử lý từng cặp
	for _, pair := range pairs {
		// Lấy program_id từ lesson (qua chapters)
		var lessonProgramID int64
		if err := db.MasterDB.Table("lessons l").
			Select("ch.program_id").
			Joins("JOIN chapters ch ON ch.id = l.chapter_id").
			Where("l.id = ? AND l.deleted_at IS NULL", pair.LessonID).
			Scan(&lessonProgramID).Error; err != nil {
			log.Printf("⚠️ Lỗi lấy program_id cho lesson_id %d: %v", pair.LessonID, err)
			continue
		}

		if lessonProgramID == 0 {
			log.Printf("⚠️ Không tìm thấy program_id cho lesson_id %d", pair.LessonID)
			continue
		}

		// Lấy tất cả course_id có cùng program_id
		var courseIDs []int64
		if err := db.MasterDB.Table("courses").
			Select("id").
			Where("program_id = ? AND deleted_at IS NULL", lessonProgramID).
			Pluck("id", &courseIDs).Error; err != nil {
			log.Printf("⚠️ Lỗi lấy course_id cho program_id %d: %v", lessonProgramID, err)
			continue
		}

		if len(courseIDs) == 0 {
			log.Printf("⚠️ Không tìm thấy course nào cho program_id %d", lessonProgramID)
			continue
		}

		// Lấy danh sách các (assessment_id, lesson_id, course_id) đã tồn tại
		var existingRecords []models.AssessmentRefLesson
		if err := db.MasterDB.
			Where("assessment_id = ? AND lesson_id = ? AND course_id IN (?)", pair.AssessmentID, pair.LessonID, courseIDs).
			Find(&existingRecords).Error; err != nil {
			log.Printf("⚠️ Lỗi kiểm tra records đã tồn tại: %v", err)
			continue
		}

		// Tạo map các course_id đã tồn tại
		existingCourseIDMap := make(map[int64]bool)
		for _, record := range existingRecords {
			if record.CourseId > 0 {
				existingCourseIDMap[record.CourseId] = true
			}
		}

		// Tạo danh sách các records cần insert (chưa tồn tại)
		now := time.Now().UTC()
		assignedBy := int64(1)
		var recordsToInsert []models.AssessmentRefLesson
		for _, courseID := range courseIDs {
			if !existingCourseIDMap[courseID] {
				recordsToInsert = append(recordsToInsert, models.AssessmentRefLesson{
					AssessmentId: pair.AssessmentID,
					LessonId:     pair.LessonID,
					CourseId:     courseID,
					AssignedBy:   &assignedBy,
					AssignedAt:   &now,
				})
			}
		}

		// Insert các records mới (sử dụng ON CONFLICT DO NOTHING để tránh lỗi duplicate)
		if len(recordsToInsert) > 0 {
			if err := db.MasterDB.Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "assessment_id"},
					{Name: "lesson_id"},
					{Name: "course_id"},
				},
				DoNothing: true,
			}).Create(&recordsToInsert).Error; err != nil {
				log.Printf("⚠️ Lỗi insert records cho assessment_id %d, lesson_id %d: %v", pair.AssessmentID, pair.LessonID, err)
				continue
			}

			totalInserted += len(recordsToInsert)
			log.Printf("  ✅ Assessment %d, Lesson %d: Thêm %d course_id mới", pair.AssessmentID, pair.LessonID, len(recordsToInsert))
		}

		totalProcessed++
	}

	log.Printf("✅ Hoàn thành! Đã xử lý %d cặp, thêm %d records", totalProcessed, totalInserted)
	return totalProcessed, totalInserted, nil
}

