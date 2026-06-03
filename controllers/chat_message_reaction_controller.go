package controllers

import (
	"be-lms/i18n"
	"be-lms/services"
	"be-lms/utils"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatMessageReactionController struct {
	service services.ChatMessageReactionService
}

func NewChatMessageReactionController(service services.ChatMessageReactionService) *ChatMessageReactionController {
	return &ChatMessageReactionController{
		service: service,
	}
}

type AddReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

type RemoveReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

// AddReaction handles POST /courses/:id/chat/messages/:messageId/reactions
func (c *ChatMessageReactionController) AddReaction(ctx *gin.Context) {
	// Get courseId from URL parameter
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get messageId from URL parameter
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to uint64
	var userID uint64
	switch v := userIDInterface.(type) {
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	case uint64:
		userID = v
	case float64:
		userID = uint64(v)
	default:
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req AddReactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Call service
	response, err := c.service.AddReaction(context.Background(), courseID, messageID, userID, req.Emoji)
	if err != nil {
		if response != nil && !response.Success {
			utils.Respond(ctx, response, err, "", http.StatusBadRequest)
		} else {
			utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		}
		return
	}

	utils.Respond(ctx, response, nil, "")
}

// RemoveReaction handles DELETE /courses/:id/chat/messages/:messageId/reactions
func (c *ChatMessageReactionController) RemoveReaction(ctx *gin.Context) {
	// Get courseId from URL parameter
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get messageId from URL parameter
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to uint64
	var userID uint64
	switch v := userIDInterface.(type) {
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	case uint64:
		userID = v
	case float64:
		userID = uint64(v)
	default:
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req RemoveReactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Call service
	response, err := c.service.RemoveReaction(context.Background(), courseID, messageID, userID, req.Emoji)
	if err != nil {
		if response != nil && !response.Success {
			utils.Respond(ctx, response, err, "", http.StatusBadRequest)
		} else {
			utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		}
		return
	}

	utils.Respond(ctx, response, nil, "")
}

// GetMessageReactions handles GET /courses/:id/chat/messages/:messageId/reactions
func (c *ChatMessageReactionController) GetMessageReactions(ctx *gin.Context) {
	// Get courseId from URL parameter
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get messageId from URL parameter
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Call service
	response, err := c.service.GetMessageReactions(context.Background(), courseID, messageID)
	if err != nil {
		if response != nil && !response.Success {
			utils.Respond(ctx, response, err, "", http.StatusBadRequest)
		} else {
			utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		}
		return
	}

	utils.Respond(ctx, response, nil, "")
}

// DeleteAllReactions handles DELETE /courses/:id/chat/messages/:messageId/reactions/all
func (c *ChatMessageReactionController) DeleteAllReactions(ctx *gin.Context) {
	// Get courseId from URL parameter
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get messageId from URL parameter
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to uint64
	var userID uint64
	switch v := userIDInterface.(type) {
	case int:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	case uint64:
		userID = v
	case float64:
		userID = uint64(v)
	default:
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized", http.StatusUnauthorized)
		return
	}

	// Call service
	response, err := c.service.DeleteAllReactions(context.Background(), courseID, messageID, userID)
	if err != nil {
		if response != nil && !response.Success {
			utils.Respond(ctx, response, err, "", http.StatusBadRequest)
		} else {
			utils.Respond(ctx, nil, err, "", http.StatusInternalServerError)
		}
		return
	}

	utils.Respond(ctx, response, nil, "")
}
