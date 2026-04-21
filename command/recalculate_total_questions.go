package command

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/services"
	"fmt"
	"strings"
)

type RecalculateTotalQuestionsCommand struct {
	assignmentType string // "homework", "exam", "exercise"
	batchSize      int
}

type AssignmentStats struct {
	Total        int      `json:"total"`
	Updated      int      `json:"updated"`
	Skipped      int      `json:"skipped"`
	Errors       int      `json:"errors"`
	ErrorDetails []string `json:"error_details,omitempty"`
}

func NewRecalculateTotalQuestionsCommand(assignmentType string) *RecalculateTotalQuestionsCommand {
	return &RecalculateTotalQuestionsCommand{
		assignmentType: assignmentType,
		batchSize:      100,
	}
}

// Execute chạy job tính lại total_questions
func (cmd *RecalculateTotalQuestionsCommand) Execute() error {
	_, err := cmd.ExecuteWithStats()
	return err
}

// ExecuteWithStats chạy job và return stats
func (cmd *RecalculateTotalQuestionsCommand) ExecuteWithStats() (*AssignmentStats, error) {
	cfg := config.LoadConfig()

	if err := db.ConnectPostgres(cfg); err != nil {
		return nil, fmt.Errorf("connect postgres error: %w", err)
	}

	fmt.Printf("🚀 Bắt đầu tính lại total_questions cho %s...\n", cmd.assignmentType)

	var stats AssignmentStats
	var err error

	switch cmd.assignmentType {
	case "homework":
		stats, err = cmd.recalculateHomeworks()
	case "exam":
		stats, err = cmd.recalculateExams()
	case "exercise":
		stats, err = cmd.recalculateExercises()
	default:
		return nil, fmt.Errorf("invalid assignment type: %s (valid: homework, exam, exercise)", cmd.assignmentType)
	}

	if err != nil {
		return &stats, err
	}

	// In kết quả
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✅ Hoàn thành tính lại total_questions cho %s\n", cmd.assignmentType)
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

// recalculateHomeworks tính lại total_questions cho homeworks
func (cmd *RecalculateTotalQuestionsCommand) recalculateHomeworks() (AssignmentStats, error) {
	stats := AssignmentStats{}
	var lastID int64 = 0

	for {
		var homeworks []models.Homework

		// Lấy batch homework (chỉ lấy những homework có cloned_questions)
		err := db.MasterDB.Table("homeworks").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(cmd.batchSize).
			Find(&homeworks).Error

		if err != nil {
			return stats, fmt.Errorf("failed to fetch homeworks: %w", err)
		}

		if len(homeworks) == 0 {
			break
		}

		// Xử lý từng homework
		for _, hw := range homeworks {
			lastID = hw.ID
			stats.Total++

			// Tính lại total_questions từ cloned_questions
			newTotalQuestions, err := services.CountHomeworkQuestionsFromCloned(hw.ID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("Homework ID %d: %v", hw.ID, err))
				continue
			}

			// Kiểm tra xem có khác với giá trị hiện tại không
			if hw.TotalQuestions != newTotalQuestions {
				// Cập nhật vào database
				err = db.MasterDB.Model(&models.Homework{}).
					Where("id = ?", hw.ID).
					Update("total_questions", newTotalQuestions).Error

				if err != nil {
					stats.Errors++
					stats.ErrorDetails = append(stats.ErrorDetails,
						fmt.Sprintf("Homework ID %d: failed to update - %v", hw.ID, err))
					continue
				}

				stats.Updated++
				fmt.Printf("✏️  Homework ID %d: %d -> %d\n", hw.ID, hw.TotalQuestions, newTotalQuestions)
			} else {
				stats.Skipped++
			}
		}

		// Progress
		fmt.Printf("📈 Progress: %d processed (updated: %d, skipped: %d, errors: %d)\n",
			stats.Total, stats.Updated, stats.Skipped, stats.Errors)
	}

	return stats, nil
}

// recalculateExams tính lại total_questions cho exams
func (cmd *RecalculateTotalQuestionsCommand) recalculateExams() (AssignmentStats, error) {
	stats := AssignmentStats{}
	var lastID int64 = 0

	for {
		var exams []models.Exam

		err := db.MasterDB.Table("exams").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(cmd.batchSize).
			Find(&exams).Error

		if err != nil {
			return stats, fmt.Errorf("failed to fetch exams: %w", err)
		}

		if len(exams) == 0 {
			break
		}

		for _, exam := range exams {
			lastID = exam.ID
			stats.Total++

			newTotalQuestions, err := services.CountExamQuestionsFromCloned(exam.ID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("Exam ID %d: %v", exam.ID, err))
				continue
			}

			if exam.TotalQuestions != newTotalQuestions {
				err = db.MasterDB.Model(&models.Exam{}).
					Where("id = ?", exam.ID).
					Update("total_questions", newTotalQuestions).Error

				if err != nil {
					stats.Errors++
					stats.ErrorDetails = append(stats.ErrorDetails,
						fmt.Sprintf("Exam ID %d: failed to update - %v", exam.ID, err))
					continue
				}

				stats.Updated++
				fmt.Printf("✏️  Exam ID %d: %d -> %d\n", exam.ID, exam.TotalQuestions, newTotalQuestions)
			} else {
				stats.Skipped++
			}
		}

		fmt.Printf("📈 Progress: %d processed (updated: %d, skipped: %d, errors: %d)\n",
			stats.Total, stats.Updated, stats.Skipped, stats.Errors)
	}

	return stats, nil
}

// recalculateExercises tính lại total_questions cho exercises
func (cmd *RecalculateTotalQuestionsCommand) recalculateExercises() (AssignmentStats, error) {
	stats := AssignmentStats{}
	var lastID int64 = 0

	for {
		var exercises []models.Exercise

		err := db.MasterDB.Table("exercises").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(cmd.batchSize).
			Find(&exercises).Error

		if err != nil {
			return stats, fmt.Errorf("failed to fetch exercises: %w", err)
		}

		if len(exercises) == 0 {
			break
		}

		for _, exercise := range exercises {
			lastID = exercise.ID
			stats.Total++

			newTotalQuestions, err := services.CountExerciseQuestionsFromCloned(exercise.ID)
			if err != nil {
				stats.Errors++
				stats.ErrorDetails = append(stats.ErrorDetails,
					fmt.Sprintf("Exercise ID %d: %v", exercise.ID, err))
				continue
			}

			if exercise.TotalQuestions != newTotalQuestions {
				err = db.MasterDB.Model(&models.Exercise{}).
					Where("id = ?", exercise.ID).
					Update("total_questions", newTotalQuestions).Error

				if err != nil {
					stats.Errors++
					stats.ErrorDetails = append(stats.ErrorDetails,
						fmt.Sprintf("Exercise ID %d: failed to update - %v", exercise.ID, err))
					continue
				}

				stats.Updated++
				fmt.Printf("✏️  Exercise ID %d: %d -> %d\n", exercise.ID, exercise.TotalQuestions, newTotalQuestions)
			} else {
				stats.Skipped++
			}
		}

		fmt.Printf("📈 Progress: %d processed (updated: %d, skipped: %d, errors: %d)\n",
			stats.Total, stats.Updated, stats.Skipped, stats.Errors)
	}

	return stats, nil
}

// RunRecalculateTotalQuestionsCommand chạy command từ CLI
func RunRecalculateTotalQuestionsCommand(assignmentType string) error {
	cmd := NewRecalculateTotalQuestionsCommand(assignmentType)
	return cmd.Execute()
}

// Uncomment để chạy standalone: go run command/recalculate_total_questions.go homework
// func main() {
// 	if len(os.Args) < 2 {
// 		fmt.Println("Usage: go run command/recalculate_total_questions.go [homework|exam|exercise]")
// 		os.Exit(1)
// 	}
//
// 	assignmentType := os.Args[1]
// 	if err := RunRecalculateTotalQuestionsCommand(assignmentType); err != nil {
// 		fmt.Printf("❌ Error: %v\n", err)
// 		os.Exit(1)
// 	}
// }
