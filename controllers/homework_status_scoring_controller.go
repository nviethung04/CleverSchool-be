package controllers

import (
	"be-cleverschool/jobs"
	"be-cleverschool/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type HomeworkStatusScoringController struct{}

func NewHomeworkStatusScoringController() *HomeworkStatusScoringController {
	return &HomeworkStatusScoringController{}
}

// SyncHomeworkStatusScoring chạy job sync homework status scoring
func (c *HomeworkStatusScoringController) SyncHomeworkStatusScoring(ctx *gin.Context) {
	startTime := time.Now()
	
	// Chạy job sync homework status scoring
	jobs.SyncHomeworkStatusScoringJob()
	
	duration := time.Since(startTime)
	
	utils.Respond(ctx, gin.H{
		"message": "Homework status scoring sync completed successfully",
		"duration": duration.String(),
		"timestamp": startTime.Format(time.RFC3339),
	}, nil, "")
}

