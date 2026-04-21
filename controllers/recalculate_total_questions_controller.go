package controllers

import (
	"be-lms/command"
	"be-lms/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecalculateTotalQuestionsController struct{}

func NewRecalculateTotalQuestionsController() *RecalculateTotalQuestionsController {
	return &RecalculateTotalQuestionsController{}
}

// RecalculateHomeworkTotalQuestions godoc
// @Summary Recalculate total questions for homeworks (async)
// @Description Dispatch job tính lại cột total_questions cho tất cả homework từ cloned_questions. Job chạy async trong background.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "job dispatched successfully"
// @Router /api/internal/command/recalculate-homework-total-questions [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateTotalQuestionsController) RecalculateHomeworkTotalQuestions(c *gin.Context) {
	// Dispatch job async
	go func() {
		cmd := command.NewRecalculateTotalQuestionsCommand("homework")
		if err := cmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate homework total questions: %v", err)
		}
	}()

	// Return ngay lập tức
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Job recalculate homework total questions has been dispatched successfully",
		"type":    "homework",
	})
}

// RecalculateExamTotalQuestions godoc
// @Summary Recalculate total questions for exams (async)
// @Description Dispatch job tính lại cột total_questions cho tất cả exam từ cloned_questions. Job chạy async trong background.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "job dispatched successfully"
// @Router /api/internal/command/recalculate-exam-total-questions [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateTotalQuestionsController) RecalculateExamTotalQuestions(c *gin.Context) {
	// Dispatch job async
	go func() {
		cmd := command.NewRecalculateTotalQuestionsCommand("exam")
		if err := cmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate exam total questions: %v", err)
		}
	}()

	// Return ngay lập tức
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Job recalculate exam total questions has been dispatched successfully",
		"type":    "exam",
	})
}

// RecalculateExerciseTotalQuestions godoc
// @Summary Recalculate total questions for exercises (async)
// @Description Dispatch job tính lại cột total_questions cho tất cả exercise từ cloned_questions. Job chạy async trong background.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "job dispatched successfully"
// @Router /api/internal/command/recalculate-exercise-total-questions [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateTotalQuestionsController) RecalculateExerciseTotalQuestions(c *gin.Context) {
	// Dispatch job async
	go func() {
		cmd := command.NewRecalculateTotalQuestionsCommand("exercise")
		if err := cmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate exercise total questions: %v", err)
		}
	}()

	// Return ngay lập tức
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Job recalculate exercise total questions has been dispatched successfully",
		"type":    "exercise",
	})
}

// RecalculateAllTotalQuestions godoc
// @Summary Recalculate total questions for all (homework, exam, exercise) (async)
// @Description Dispatch job tính lại cột total_questions cho tất cả homework, exam, exercise từ cloned_questions. Job chạy async trong background.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "jobs dispatched successfully"
// @Router /api/internal/command/recalculate-all-total-questions [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateTotalQuestionsController) RecalculateAllTotalQuestions(c *gin.Context) {
	// Dispatch tất cả jobs async
	go func() {
		// Homework
		homeworkCmd := command.NewRecalculateTotalQuestionsCommand("homework")
		if err := homeworkCmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate homework total questions: %v", err)
		}

		// Exam
		examCmd := command.NewRecalculateTotalQuestionsCommand("exam")
		if err := examCmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate exam total questions: %v", err)
		}

		// Exercise
		exerciseCmd := command.NewRecalculateTotalQuestionsCommand("exercise")
		if err := exerciseCmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate exercise total questions: %v", err)
		}
	}()

	// Return ngay lập tức
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Jobs recalculate all total questions (homework, exam, exercise) have been dispatched successfully",
		"types":   []string{"homework", "exam", "exercise"},
	})
}

