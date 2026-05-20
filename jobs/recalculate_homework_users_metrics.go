package jobs

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"fmt"
	"time"
)

// RecalculateHomeworkUsersMetricsJob tính lại metrics cho homework_users theo điều kiện
// Áp dụng logic y hệt như submit homework
func RecalculateHomeworkUsersMetricsJob(startDate, endDate *time.Time, onlyHomeworkZeroScore bool) (totalProcessed int64, totalUpdated int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in RecalculateHomeworkUsersMetricsJob: %v", r)
		}
	}()

	config.Log.Info("Starting RecalculateHomeworkUsersMetricsJob")

	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return 0, 0, fmt.Errorf("database not connected")
	}

	startTime := time.Now()

	// Khởi tạo services
	homeworkRepo := repositories.NewHomeworkRepository()
	homeworkUserService := services.NewHomeworkUserService(
		repositories.NewHomeworkUserRepository(),
		homeworkRepo,
		services.NewClonedQuestionService(repositories.NewClonedQuestionRepository()),
		repositories.NewSaveScoreManualScoringRepository(),
	)
	homeworkCalculationService := services.NewHomeworkCalculationService(homeworkUserService)
	userStarExpService := services.NewUserStarExpService()

	// Query homework_users với điều kiện
	query := db.MasterDB.Table("homework_users")

	// Filter theo start_date/end_date (filter theo created_at)
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		// Thêm 1 ngày để bao gồm cả ngày end_date
		endDateInclusive := endDate.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endDateInclusive)
	}

	// Filter chỉ những homework có ratio = 0 nếu only_homework_zero_score = true
	if onlyHomeworkZeroScore {
		query = query.Where("ratio = ?", 0)
	}

	// Lấy tất cả records
	var homeworkUsers []models.HomeworkUser
	if err := query.Find(&homeworkUsers).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to fetch homework_users: %w", err)
	}

	config.Log.Infof("Found %d homework_users to process", len(homeworkUsers))

	var processed, updated int64

	// Xử lý từng homework_user
	for _, hu := range homeworkUsers {
		processed++

		// Lấy lesson_id từ homework_user record
		lessonID := hu.LessonID
		if lessonID == 0 {
			config.Log.Warnf("HomeworkUser ID %d (homework_id=%d, user_id=%d) has lesson_id = 0, skipping", hu.ID, hu.HomeworkID, hu.UserID)
			continue
		}

		// Lấy giá trị cũ để tính change_star và change_exp
		oldStar := int64(hu.Star)
		oldExp := hu.Exp

		// Tính toán lại metrics qua CalculateHomeworkMetrics (giống submit homework)
		calculationResult, err := homeworkCalculationService.CalculateHomeworkMetrics(hu.HomeworkID, hu.UserID, lessonID)
		if err != nil {
			config.Log.Errorf("HomeworkUser ID %d: CalculateHomeworkMetrics failed - %v", hu.ID, err)
			continue
		}

		// Lưu score
		if err := homeworkUserService.SaveHomeworkScoreService(hu.HomeworkID, hu.UserID, calculationResult.Score); err != nil {
			config.Log.Errorf("HomeworkUser ID %d: SaveHomeworkScoreService failed - %v", hu.ID, err)
			continue
		}

		// Lưu ratio
		if err := homeworkUserService.SaveHomeworkRatioService(hu.HomeworkID, hu.UserID, calculationResult.Ratio); err != nil {
			config.Log.Errorf("HomeworkUser ID %d: SaveHomeworkRatioService failed - %v", hu.ID, err)
			continue
		}

		// Lưu exp
		if err := homeworkUserService.SaveHomeworkExpService(hu.HomeworkID, hu.UserID, calculationResult.Exp); err != nil {
			config.Log.Errorf("HomeworkUser ID %d: SaveHomeworkExpService failed - %v", hu.ID, err)
			continue
		}

		// Lưu star
		if err := homeworkUserService.SaveHomeworkStarService(hu.HomeworkID, hu.UserID, calculationResult.Star); err != nil {
			config.Log.Errorf("HomeworkUser ID %d: SaveHomeworkStarService failed - %v", hu.ID, err)
			continue
		}

		// Tính change_star và change_exp (giống submit homework)
		changeStar := calculationResult.Star - oldStar
		changeExp := calculationResult.Exp - oldExp

		// Lưu tổng star và exp vào bảng user_star_exp (giống submit homework)
		note := fmt.Sprintf("homework %d (recalculate)", hu.HomeworkID)
		if err := userStarExpService.UpdateStarAndExp(hu.UserID, changeStar, changeExp, note); err != nil {
			config.Log.Warnf("HomeworkUser ID %d: UpdateStarAndExp failed - %v", hu.ID, err)
			// Không dừng flow, chỉ log warning
		}

		// Kiểm tra và lưu has_manual_scoring (giống submit homework)
		hasManualScoring, err := homeworkUserService.CheckHomeworkHasManualScoringFromClonedQuestions(hu.HomeworkID)
		if err != nil {
			config.Log.Errorf("HomeworkUser ID %d: CheckHomeworkHasManualScoringFromClonedQuestions failed - %v", hu.ID, err)
			continue
		}

		if err := homeworkUserService.SaveHomeworkHasManualScoringService(hu.HomeworkID, hu.UserID, hasManualScoring); err != nil {
			config.Log.Errorf("HomeworkUser ID %d: SaveHomeworkHasManualScoringService failed - %v", hu.ID, err)
			continue
		}

		// Cập nhật status_scoring (giống submit homework)
		if err := homeworkUserService.UpdateHomeworkStatusScoringService(hu.HomeworkID, hu.UserID); err != nil {
			config.Log.Warnf("HomeworkUser ID %d: UpdateHomeworkStatusScoringService failed - %v", hu.ID, err)
			// Không dừng flow, chỉ log warning
		}

		// Cập nhật updated_at = MAX(created_at) từ homework_user_questions theo bộ 3 (homework_id, user_id, lesson_id)
		var maxCreatedAt *time.Time
		err = db.MasterDB.Table("homework_user_questions").
			Select("MAX(created_at) as max_created_at").
			Where("homework_id = ? AND user_id = ? AND lesson_id = ?", hu.HomeworkID, hu.UserID, lessonID).
			Scan(&maxCreatedAt).Error
		
		var finalUpdatedAt time.Time
		if err != nil || maxCreatedAt == nil || maxCreatedAt.IsZero() {
			// Nếu không có record hoặc lỗi, giữ nguyên updated_at hiện tại của homework_users
			finalUpdatedAt = hu.UpdatedAt
			if finalUpdatedAt.IsZero() {
				finalUpdatedAt = time.Now()
			}
			config.Log.Warnf("HomeworkUser ID %d: No records in homework_user_questions or error, keeping current updated_at", hu.ID)
		} else {
			finalUpdatedAt = *maxCreatedAt
		}

		// Update updated_at của homework_users
		err = db.MasterDB.Model(&models.HomeworkUser{}).
			Where("homework_id = ? AND user_id = ? AND lesson_id = ?", hu.HomeworkID, hu.UserID, lessonID).
			Update("updated_at", finalUpdatedAt).Error
		if err != nil {
			config.Log.Warnf("HomeworkUser ID %d: Failed to update updated_at - %v", hu.ID, err)
			// Không dừng flow, chỉ log warning
		}

		updated++
		config.Log.Infof("HomeworkUser ID %d: updated - score=%.2f, ratio=%.2f, exp=%.2f, star=%d, is_completed=%v, updated_at=%v",
			hu.ID, calculationResult.Score, calculationResult.Ratio, calculationResult.Exp, calculationResult.Star, calculationResult.IsCompleted, finalUpdatedAt)
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished RecalculateHomeworkUsersMetricsJob in %v, processed=%d, updated=%d", duration, processed, updated)

	return processed, updated, nil
}

