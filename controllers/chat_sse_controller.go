package controllers

import (
	"be-lms/services"
	"be-lms/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatSSEController struct {
	pubsubService services.ChatPubSubService
}

func NewChatSSEController() *ChatSSEController {
	return &ChatSSEController{
		pubsubService: services.NewChatPubSubService(),
	}
}

// StreamCourseEvents handles SSE connection for course chat events
// GET /api/manage/courses/:id/chat/events?token=jwt_token
func (c *ChatSSEController) StreamCourseEvents(ctx *gin.Context) {
	// Get token from query parameter for SSE compatibility
	tokenStr := ctx.Query("token")
	if tokenStr == "" {
		tokenStr = ctx.GetHeader("Token")
	}

	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
		return
	}

	// Validate token and get user ID
	claims, err := utils.CheckTokenWithoutContext(tokenStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	userID := uint64(claims.UserID)

	// Get course ID
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Subscribe to course events
	messageChan, cleanup, err := c.pubsubService.SubscribeToCourse(courseID)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", err.Error())
		ctx.Writer.Flush()
		return
	}
	defer cleanup()

	// Send initial connection success event
	fmt.Fprintf(ctx.Writer, "event: connected\ndata: {\"course_id\": %d, \"user_id\": %d, \"timestamp\": \"%s\"}\n\n",
		courseID, userID, time.Now().Format(time.RFC3339))
	ctx.Writer.Flush()

	// Handle client disconnect
	clientGone := ctx.Writer.CloseNotify()
	requestCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	// Send heartbeat every 30 seconds
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-clientGone:
			fmt.Printf("SSE client disconnected for course %d\n", courseID)
			return

		case <-requestCtx.Done():
			fmt.Printf("SSE context cancelled for course %d\n", courseID)
			return

		case <-heartbeat.C:
			// Send heartbeat
			fmt.Fprintf(ctx.Writer, "event: heartbeat\ndata: {\"timestamp\": \"%s\"}\n\n",
				time.Now().Format(time.RFC3339))
			ctx.Writer.Flush()

		case msg, ok := <-messageChan:
			if !ok {
				fmt.Fprintf(ctx.Writer, "event: error\ndata: Message channel closed\n\n")
				ctx.Writer.Flush()
				return
			}

			if msg != nil {
				// Convert message to JSON
				jsonData, err := json.Marshal(msg)
				if err != nil {
					fmt.Printf("Failed to marshal SSE message: %v\n", err)
					continue
				}

				// Send event based on message type
				eventType := msg.Type
				if eventType == "" {
					eventType = "message"
				}

				fmt.Fprintf(ctx.Writer, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
				ctx.Writer.Flush()
			}
		}
	}
}

// StreamUserEvents handles SSE connection for user-specific notifications
// GET /api/manage/users/events?token=jwt_token
func (c *ChatSSEController) StreamUserEvents(ctx *gin.Context) {
	// Get token from query parameter for SSE compatibility
	tokenStr := ctx.Query("token")
	if tokenStr == "" {
		tokenStr = ctx.GetHeader("Token")
	}

	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
		return
	}

	// Validate token and get user ID
	claims, err := utils.CheckTokenWithoutContext(tokenStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	userID := uint64(claims.UserID)

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Subscribe to user notifications
	messageChan, cleanup, err := c.pubsubService.SubscribeToUser(userID)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", err.Error())
		ctx.Writer.Flush()
		return
	}
	defer cleanup()

	// Send initial connection success event
	fmt.Fprintf(ctx.Writer, "event: connected\ndata: {\"user_id\": %d, \"timestamp\": \"%s\"}\n\n",
		userID, time.Now().Format(time.RFC3339))
	ctx.Writer.Flush()

	// Handle client disconnect
	clientGone := ctx.Writer.CloseNotify()
	requestCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	// Send heartbeat every 30 seconds
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-clientGone:
			fmt.Printf("SSE client disconnected for user %d\n", userID)
			return

		case <-requestCtx.Done():
			fmt.Printf("SSE context cancelled for user %d\n", userID)
			return

		case <-heartbeat.C:
			// Send heartbeat
			fmt.Fprintf(ctx.Writer, "event: heartbeat\ndata: {\"timestamp\": \"%s\"}\n\n",
				time.Now().Format(time.RFC3339))
			ctx.Writer.Flush()

		case msg, ok := <-messageChan:
			if !ok {
				fmt.Fprintf(ctx.Writer, "event: error\ndata: Message channel closed\n\n")
				ctx.Writer.Flush()
				return
			}

			if msg != nil {
				// Convert message to JSON
				jsonData, err := json.Marshal(msg)
				if err != nil {
					fmt.Printf("Failed to marshal SSE message: %v\n", err)
					continue
				}

				// Send notification event
				fmt.Fprintf(ctx.Writer, "event: notification\ndata: %s\n\n", string(jsonData))
				ctx.Writer.Flush()
			}
		}
	}
}

// StreamMultipleCourseEvents handles SSE connection for multiple courses
// GET /api/manage/courses/events?course_ids=1,2,3&token=jwt_token
func (c *ChatSSEController) StreamMultipleCourseEvents(ctx *gin.Context) {
	// Get token from query parameter for SSE compatibility
	tokenStr := ctx.Query("token")
	if tokenStr == "" {
		tokenStr = ctx.GetHeader("Token")
	}

	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
		return
	}

	// Validate token and get user ID
	claims, err := utils.CheckTokenWithoutContext(tokenStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	userID := uint64(claims.UserID)

	// Parse course IDs from query parameter
	courseIDsStr := ctx.Query("course_ids")
	if courseIDsStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "course_ids parameter is required"})
		return
	}

	courseIDStrs := strings.Split(courseIDsStr, ",")
	courseIDs := make([]uint64, 0, len(courseIDStrs))

	for _, idStr := range courseIDStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		courseID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid course ID: %s", idStr)})
			return
		}
		courseIDs = append(courseIDs, courseID)
	}

	if len(courseIDs) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid course IDs provided"})
		return
	}

	// Set SSE headers
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Subscribe to multiple course events
	messageChan, cleanup, err := c.pubsubService.SubscribeToMultipleCourses(courseIDs)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", err.Error())
		ctx.Writer.Flush()
		return
	}
	defer cleanup()

	// Send initial connection success event
	courseIDsJson, _ := json.Marshal(courseIDs)
	fmt.Fprintf(ctx.Writer, "event: connected\ndata: {\"user_id\": %d, \"course_ids\": %s, \"timestamp\": \"%s\"}\n\n",
		userID, string(courseIDsJson), time.Now().Format(time.RFC3339))
	ctx.Writer.Flush()

	// Handle client disconnect
	clientGone := ctx.Writer.CloseNotify()
	requestCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	// Send heartbeat every 30 seconds
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-clientGone:
			fmt.Printf("SSE client disconnected for user %d, courses %v\n", userID, courseIDs)
			return

		case <-requestCtx.Done():
			fmt.Printf("SSE context cancelled for user %d, courses %v\n", userID, courseIDs)
			return

		case <-heartbeat.C:
			// Send heartbeat
			fmt.Fprintf(ctx.Writer, "event: heartbeat\ndata: {\"timestamp\": \"%s\"}\n\n",
				time.Now().Format(time.RFC3339))
			ctx.Writer.Flush()

		case msg, ok := <-messageChan:
			if !ok {
				fmt.Fprintf(ctx.Writer, "event: error\ndata: Message channel closed\n\n")
				ctx.Writer.Flush()
				return
			}

			if msg != nil {
				// Convert message to JSON
				jsonData, err := json.Marshal(msg)
				if err != nil {
					fmt.Printf("Failed to marshal SSE message: %v\n", err)
					continue
				}

				// Send event based on message type
				eventType := msg.Type
				if eventType == "" {
					eventType = "message"
				}

				fmt.Fprintf(ctx.Writer, "event: %s\ndata: %s\n\n", eventType, string(jsonData))
				ctx.Writer.Flush()
			}
		}
	}
}

// GetActiveChannels returns active chat channels for monitoring
// GET /api/manage/chat/channels
func (c *ChatSSEController) GetActiveChannels(ctx *gin.Context) {
	channels, err := c.pubsubService.GetActiveChannels()
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(ctx, gin.H{"channels": channels}, nil, "", http.StatusOK)
}

// GetChannelStats returns subscriber statistics for channels
// GET /api/manage/chat/channels/stats
func (c *ChatSSEController) GetChannelStats(ctx *gin.Context) {
	channels, err := c.pubsubService.GetActiveChannels()
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		return
	}

	stats := make(map[string]int64)
	for _, channel := range channels {
		count, err := c.pubsubService.GetChannelSubscribers(channel)
		if err != nil {
			fmt.Printf("Failed to get subscriber count for channel %s: %v\n", channel, err)
			count = 0
		}
		stats[channel] = count
	}

	utils.Respond(ctx, gin.H{
		"channels":          channels,
		"subscriber_counts": stats,
		"total_channels":    len(channels),
	}, nil, "", http.StatusOK)
}
