package controllers

import (
	"be-lms/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScormController struct {
	scormService *services.ScormService
}

func NewScormController() *ScormController {
	return &ScormController{
		scormService: services.NewScormService(),
	}
}

// LaunchScorm launches a SCORM activity for a user
func (c *ScormController) LaunchScorm(ctx *gin.Context) {
	activityIDStr := ctx.Param("id")
	userIDStr := ctx.Query("userId")

	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	activityID, err := strconv.ParseUint(activityIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Create new attempt
	attempt, err := c.scormService.CreateScormAttempt(uint(activityID), uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get activity details
	activity, err := c.scormService.GetScormActivity(uint(activityID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Redirect to SCORM viewer with parameters
	viewerURL := "/scorm-viewer?" +
		"version=" + activity.Version +
		"&attemptId=" + attempt.ID +
		"&courseId=" + strconv.FormatUint(uint64(activityID), 10) +
		"&userId=" + userIDStr +
		"&contentUrl=" + activity.LaunchURL

	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"attemptId": attempt.ID,
		"viewerUrl": viewerURL,
		"activity": gin.H{
			"id":      activity.ID,
			"title":   activity.Title,
			"version": activity.Version,
		},
	})
}

// GetScormData gets CMI data for a SCORM attempt
func (c *ScormController) GetScormData(ctx *gin.Context) {
	attemptID := ctx.Query("attemptId")
	element := ctx.Query("element")

	if attemptID == "" || element == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "attemptId and element are required"})
		return
	}

	value, err := c.scormService.GetScormCMI(attemptID, element)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return empty string if no value (SCORM standard)
	ctx.String(http.StatusOK, value)
}

// SetScormData sets CMI data for a SCORM attempt
func (c *ScormController) SetScormData(ctx *gin.Context) {
	var payload struct {
		AttemptID string `json:"attemptId" binding:"required"`
		Element   string `json:"element" binding:"required"`
		Value     string `json:"value"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	err := c.scormService.SetScormCMI(payload.AttemptID, payload.Element, payload.Value)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// CommitScormData commits pending SCORM data
func (c *ScormController) CommitScormData(ctx *gin.Context) {
	var payload struct {
		AttemptID string `json:"attemptId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// For now, just return success
	// In a real implementation, you might want to flush data to persistent storage
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// TerminateScormSession terminates a SCORM session
func (c *ScormController) TerminateScormSession(ctx *gin.Context) {
	var payload struct {
		AttemptID string `json:"attemptId" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Mark attempt as completed
	err := c.scormService.CompleteScormAttempt(payload.AttemptID, "completed")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// GetScormAttempt gets SCORM attempt details
func (c *ScormController) GetScormAttempt(ctx *gin.Context) {
	attemptID := ctx.Param("id")

	attempt, err := c.scormService.GetScormAttempt(attemptID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         attempt.ID,
		"activityId": attempt.ActivityID,
		"userId":     attempt.UserID,
		"status":     attempt.Status,
		"startTime":  attempt.StartTime,
		"endTime":    attempt.EndTime,
		"createdAt":  attempt.CreatedAt,
		"updatedAt":  attempt.UpdatedAt,
	})
}

// ListScormActivities lists available SCORM activities
func (c *ScormController) ListScormActivities(ctx *gin.Context) {
	// This would typically implement pagination and filtering
	// For now, return a simple response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "SCORM activities listing endpoint",
		"note":    "Implement pagination and filtering as needed",
	})
}
