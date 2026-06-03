package controllers

import (
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/services"
	"be-lms/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatReplyController struct {
	service services.ChatReplyService
}

func NewChatReplyController(service services.ChatReplyService) *ChatReplyController {
	return &ChatReplyController{
		service: service,
	}
}

// SendReply handles POST /courses/:id/chat/messages/:messageId/replies
// Unified method that handles replies with media attachments (JSON only)
func (c *ChatReplyController) SendReply(ctx *gin.Context) {
	fmt.Printf("=== DEBUG: SendReply START ===\n")

	// Get course ID
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Get message ID (parent message)
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Get user ID from context
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized")
		return
	}

	var userID uint64
	switch v := userIDInterface.(type) {
	case int:
		userID = uint64(v)
	case uint64:
		userID = v
	case int64:
		userID = uint64(v)
	default:
		fmt.Printf("DEBUG: userID có type không hỗ trợ: %T, value: %+v\n", userIDInterface, userIDInterface)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Parse request body (JSON only, same as SendMessageWithMedias)
	var req dto.SendChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Lỗi bind JSON: %v\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	fmt.Printf("DEBUG: SendReply - courseID: %d, messageID: %d, userID: %d, content: %s, media_ids: %v\n",
		courseID, messageID, userID, req.Content, req.MediaIDs)

	// Call service
	result, err := c.service.SendReplyWithMedias(ctx.Request.Context(), courseID, userID, messageID, req)
	if err != nil {
		fmt.Printf("DEBUG: Lỗi service.SendReplyWithMedias: %v\n", err)
		utils.Respond(ctx, nil, err, "")
		return
	}

	fmt.Printf("DEBUG: SendReply thành công với ID: %d\n", result.ID)
	utils.Respond(ctx, result, nil, "")
	fmt.Printf("=== DEBUG: SendReply END ===\n")
}

// GetReplies handles GET /courses/:id/chat/messages/:messageId/replies
func (c *ChatReplyController) GetReplies(ctx *gin.Context) {
	// Get message ID
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get pagination params
	page := 1
	limit := 20

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Call service
	replies, total, err := c.service.GetRepliesByMessageID(ctx.Request.Context(), messageID, page, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusBadRequest)
		return
	}

	// Build response
	response := gin.H{
		"replies": replies,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	utils.Respond(ctx, response, nil, "", http.StatusOK)
}

// GetMessageWithReplies handles GET /courses/:id/chat/messages/:messageId/with-replies
func (c *ChatReplyController) GetMessageWithReplies(ctx *gin.Context) {
	// Get message ID
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Call service
	result, err := c.service.GetMessageWithReplies(ctx.Request.Context(), messageID)
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, result, nil,"", http.StatusOK)
}

// DeleteReply handles DELETE /courses/:id/chat/messages/:messageId/replies/:replyId
func (c *ChatReplyController) DeleteReply(ctx *gin.Context) {
	// Get reply ID
	replyIDStr := ctx.Param("replyId")
	replyID, err := strconv.ParseUint(replyIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New("Unauthorized"), "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userIDInt uint64
	switch v := userID.(type) {
	case int:
		userIDInt = uint64(v)
	case int64:
		userIDInt = uint64(v)
	case uint64:
		userIDInt = v
	default:
		utils.Respond(ctx, nil, errors.New("Invalid user ID type"), "Invalid user ID type", http.StatusBadRequest)
		return
	}

	// Call service
	err = c.service.DeleteReply(ctx.Request.Context(), replyID, userIDInt)
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, gin.H{"message": "Reply deleted successfully"}, nil, "", http.StatusOK)
}
