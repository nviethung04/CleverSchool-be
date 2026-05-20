package controllers

import (
	"be-Clever School/command"
	"be-Clever School/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecalculateHomeworkUsersController struct{}

func NewRecalculateHomeworkUsersController() *RecalculateHomeworkUsersController {
	return &RecalculateHomeworkUsersController{}
}

// RecalculateHomeworkUsersData godoc
// @Summary Recalculate homework_users data (async)
// @Description Dispatch job chuẩn hóa dữ liệu trong bảng homework_users: cập nhật lại questions_completed, score, ratio. Job chạy async trong background.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "job dispatched successfully"
// @Router /api/internal/command/recalculate-homework-users-data [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateHomeworkUsersController) RecalculateHomeworkUsersData(c *gin.Context) {
	// Dispatch job async
	go func() {
		cmd := command.NewRecalculateHomeworkUsersDataCommand()
		if err := cmd.Execute(); err != nil {
			config.Log.Errorf("Failed to recalculate homework users data: %v", err)
		}
	}()

	// Return ngay lập tức
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Job recalculate homework users data has been dispatched successfully",
		"type":    "homework_users",
	})
}

