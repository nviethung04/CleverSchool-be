package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/jobs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecalculateStudyReportsStarController struct{}

func NewRecalculateStudyReportsStarController() *RecalculateStudyReportsStarController {
	return &RecalculateStudyReportsStarController{}
}

// RecalculateStudyReportsStar godoc
// @Summary Recalculate total_star and avg_star for all study_reports (async)
// @Description Dispatch job tính lại total_star và avg_star cho tất cả study_reports. Job chạy async, xử lý theo chunk 100.
// @Tags Internal Command
// @Accept json
// @Produce json
// @Success 202 {object} map[string]interface{} "job dispatched successfully"
// @Router /api/internal/command/recalculate-study-reports-star [post]
// @Security ApiKeyAuth
func (ctrl *RecalculateStudyReportsStarController) RecalculateStudyReportsStar(c *gin.Context) {
	go func() {
		_, _, err := jobs.RecalculateStudyReportsStarJob()
		if err != nil {
			config.Log.Errorf("Failed to recalculate study reports star: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "dispatched",
		"message": "Job recalculate study reports star has been dispatched successfully",
		"type":    "study_reports",
	})
}

