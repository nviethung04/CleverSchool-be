package controllers

import (
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MeetingNotificationController struct {
	repo repositories.MeetingNotificationRepository
}

func NewMeetingNotificationController() *MeetingNotificationController {
	return &MeetingNotificationController{
		repo: repositories.NewMeetingNotificationRepository(),
	}
}

// ListMyNotifications returns meeting notifications for the current user
// @Summary List meeting notifications
// @Description Get meeting notifications for the current user
// @Tags Meeting Notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param page query int false "Page" default(1)
// @Success 200 {object} map[string]interface{} "notifications list"
// @Router /microsoft/meeting-notifications [get]
func (c *MeetingNotificationController) ListMyNotifications(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	notifs, err := c.repo.ListByUserAndProvider(uint(userID), "microsoft", limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"notifications": notifs,
		"page":          page,
		"limit":         limit,
	})
}

// MarkRead marks a notification as read
// @Summary Mark notification as read
// @Description Mark a meeting notification as read
// @Tags Meeting Notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]interface{} "success"
// @Router /microsoft/meeting-notifications/{id}/read [put]
func (c *MeetingNotificationController) MarkRead(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notifID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := c.repo.MarkRead(uint(notifID), uint(userID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
}

// CountUnread returns count of unread notifications
// @Summary Count unread notifications
// @Description Get count of unread meeting notifications
// @Tags Meeting Notifications
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "count"
// @Router /microsoft/meeting-notifications/unread-count [get]
func (c *MeetingNotificationController) CountUnread(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := c.repo.CountUnread(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"count": count})
}
