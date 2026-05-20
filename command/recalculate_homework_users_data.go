package command

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"fmt"
	"strings"
)

type RecalculateHomeworkUsersDataCommand struct {
	batchSize int
}

type HomeworkUsersRecalcStats struct {
	Total        int      `json:"total"`
	Updated      int      `json:"updated"`
	Skipped      int      `json:"skipped"`
	Errors       int      `json:"errors"`
	ErrorDetails []string `json:"error_details,omitempty"`
}

func NewRecalculateHomeworkUsersDataCommand() *RecalculateHomeworkUsersDataCommand {
	return &RecalculateHomeworkUsersDataCommand{
		batchSize: 100,
	}
}

// Execute chạy job chuẩn hóa dữ liệu homework_users
func (cmd *RecalculateHomeworkUsersDataCommand) Execute() error {
	_, err := cmd.ExecuteWithStats()
	return err
}

// ExecuteWithStats chạy job và return stats
func (cmd *RecalculateHomeworkUsersDataCommand) ExecuteWithStats() (*HomeworkUsersRecalcStats, error) {
	cfg := config.LoadConfig()

	if err := db.ConnectPostgres(cfg); err != nil {
		return nil, fmt.Errorf("connect postgres error: %w", err)
	}

	fmt.Println("🚀 Bắt đầu chuẩn hóa dữ liệu homework_users...")

	stats, err := cmd.recalculateHomeworkUsers()
	if err != nil {
		return &stats, err
	}

	// In kết quả
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ Hoàn thành chuẩn hóa dữ liệu homework_users")
	fmt.Printf("📊 Tổng số: %d\n", stats.Total)
	fmt.Printf("✏️  Đã cập nhật: %d\n", stats.Updated)
	fmt.Printf("⏭️  Bỏ qua (không thay đổi): %d\n", stats.Skipped)
	fmt.Printf("❌ Lỗi: %d\n", stats.Errors)

	if len(stats.ErrorDetails) > 0 {
		fmt.Println("\n⚠️  Chi tiết lỗi:")
		for _, errMsg := range stats.ErrorDetails {
			fmt.Printf("   - %s\n", errMsg)
		}
	}
	fmt.Println(strings.Repeat("=", 60))

	return &stats, nil
}

// recalculateHomeworkUsers chuẩn hóa dữ liệu cho tất cả homework_users
func (cmd *RecalculateHomeworkUsersDataCommand) recalculateHomeworkUsers() (HomeworkUsersRecalcStats, error) {
	stats := HomeworkUsersRecalcStats{}
	var lastID int64 = 0

	// Khởi tạo services
	skipQuestionService := services.NewSkipQuestionService()
	homeworkRepo := repositories.NewHomeworkRepository()
	homeworkUserService := services.NewHomeworkUserService(
		repositories.NewHomeworkUserRepository(),
		homeworkRepo,
		services.NewClonedQuestionService(repositories.NewClonedQuestionRepository()),
		repositories.NewSaveScoreManualScoringRepository(),
	)

	for {
		var homeworkUsers []models.HomeworkUser

		// Lấy batch homework_users
		err := db.MasterDB.Table("homework_users").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(cmd.batchSize).
			Find(&homeworkUsers).Error

		if err != nil {
			return stats, fmt.Errorf("failed to fetch homework_users: %w", err)
		}

		if len(homeworkUsers) == 0 {
			break
		}

		// Xử lý từng homework_user
		for _, hu := range homeworkUsers {
			lastID = hu.ID
			stats.Total++

			// Lưu giá trị cũ để so sánh và giữ nguyên updated_at
			oldQuestionsCompleted := hu.QuestionsCompleted
			oldScore := hu.Score
			oldRatio := hu.Ratio
			oldUpdatedAt := hu.UpdatedAt

			// Bước 1: Cập nhật questions_completed dùng CheckSubmitHomework
			checkResult, err := skipQuestionService.CheckSubmitHomework(hu.HomeworkID, hu.UserID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d (homework_id=%d, user_id=%d): CheckSubmitHomework failed - %v",
						hu.ID, hu.HomeworkID, hu.UserID, err))
				continue
			}

			// Lấy questions_completed mới sau khi CheckSubmitHomework cập nhật
			var newQuestionsCompleted int64
			err = db.MasterDB.Table("homework_users").
				Select("questions_completed").
				Where("homework_id = ? AND user_id = ?", hu.HomeworkID, hu.UserID).
				Scan(&newQuestionsCompleted).Error
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d: failed to get updated questions_completed - %v", hu.ID, err))
				continue
			}

			// Bước 2: Tính lại score
			newScore, err := homeworkUserService.CalculateHomeworkScoreService(hu.HomeworkID, hu.UserID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d: CalculateHomeworkScoreService failed - %v", hu.ID, err))
				continue
			}

			// Lưu score mới
			err = homeworkUserService.SaveHomeworkScoreService(hu.HomeworkID, hu.UserID, newScore)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d: SaveHomeworkScoreService failed - %v", hu.ID, err))
				continue
			}

			// Bước 3: Tính lại ratio
			newRatio, err := homeworkUserService.CalculateHomeworkRatioService(hu.HomeworkID, hu.UserID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d: CalculateHomeworkRatioService failed - %v", hu.ID, err))
				continue
			}

			// Lưu ratio mới
			err = homeworkUserService.SaveHomeworkRatioService(hu.HomeworkID, hu.UserID, newRatio)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("HomeworkUser ID %d: SaveHomeworkRatioService failed - %v", hu.ID, err))
				continue
			}

			// Bước 4: Cập nhật status_scoring
			err = homeworkUserService.UpdateHomeworkStatusScoringService(hu.HomeworkID, hu.UserID)
			if err != nil {
				// Log error nhưng không tính là failed
				config.Log.Warnf("HomeworkUser ID %d: UpdateHomeworkStatusScoringService failed - %v", hu.ID, err)
			}

			// Bước 5: QUAN TRỌNG - Restore lại updated_at cũ
			err = db.MasterDB.Model(&models.HomeworkUser{}).
				Where("homework_id = ? AND user_id = ?", hu.HomeworkID, hu.UserID).
				UpdateColumn("updated_at", oldUpdatedAt).Error
			if err != nil {
				config.Log.Warnf("HomeworkUser ID %d: Failed to restore updated_at - %v", hu.ID, err)
			}

			// Kiểm tra xem có thay đổi không
			hasChanged := oldQuestionsCompleted != newQuestionsCompleted ||
				oldScore != newScore ||
				oldRatio != newRatio

			if hasChanged {
				stats.Updated++
				fmt.Printf("✏️  HomeworkUser ID %d: questions=%d->%d, score=%.2f->%.2f, ratio=%.2f->%.2f [updated_at preserved]\n",
					hu.ID, oldQuestionsCompleted, newQuestionsCompleted, oldScore, newScore, oldRatio, newRatio)
			} else {
				stats.Skipped++
			}

			// Log ignored questions if any
			if checkResult.SkippedCount > 0 {
				config.Log.Infof("HomeworkUser ID %d has %d skipped questions", hu.ID, checkResult.SkippedCount)
			}
		}

		// Progress
		fmt.Printf("📈 Progress: %d processed (updated: %d, skipped: %d, errors: %d)\n",
			stats.Total, stats.Updated, stats.Skipped, stats.Errors)
	}

	return stats, nil
}

// RunRecalculateHomeworkUsersDataCommand chạy command từ CLI
func RunRecalculateHomeworkUsersDataCommand() error {
	cmd := NewRecalculateHomeworkUsersDataCommand()
	return cmd.Execute()
}


