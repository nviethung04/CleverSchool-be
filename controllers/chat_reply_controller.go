package controllers

import (
	"be-cleverschool/dto"
	"be-cleverschool/i18n"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"errors"
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

func (c *ChatReplyController) SendReply(ctx *gin.Context) {
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userId := utils.GetCurrentUserId(ctx)

	var req dto.SendChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	result, err := c.service.SendReplyWithMedias(ctx.Request.Context(), courseID, uint64(userId), messageID, req)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, result, nil, "")
}

func (c *ChatReplyController) GetReplies(ctx *gin.Context) {
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

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

	replies, total, err := c.service.GetRepliesByMessageID(ctx.Request.Context(), messageID, page, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusBadRequest)
		return
	}

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

func (c *ChatReplyController) GetMessageWithReplies(ctx *gin.Context) {
	messageIDStr := ctx.Param("messageId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	result, err := c.service.GetMessageWithReplies(ctx.Request.Context(), messageID)
	if err != nil {
		utils.Respond(ctx, nil, err, "", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, result, nil, "", http.StatusOK)
}

