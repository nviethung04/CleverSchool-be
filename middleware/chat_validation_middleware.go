package middleware

import (
	"be-Clever School/i18n"
	"be-Clever School/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidateCourseAndMessage middleware validates courseId and messageId parameters
func ValidateCourseAndMessage() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Validate courseId
		courseIDStr := c.Param("courseId")
		if courseIDStr == "" {
			courseIDStr = c.Param("id") // fallback to :id parameter
		}

		if courseIDStr == "" {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.course_id_required")), "messages.course_id_required", http.StatusBadRequest)
			c.Abort()
			return
		}

		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
			c.Abort()
			return
		}

		// Validate messageId if present
		messageIDStr := c.Param("messageId")
		if messageIDStr != "" {
			messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
			if err != nil {
				utils.Respond(c, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
				c.Abort()
				return
			}
			c.Set("messageID", messageID)
		}

		c.Set("courseID", courseID)
		c.Next()
	})
}

// ValidateEmoji middleware validates emoji format and content
func ValidateEmoji() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		type EmojiRequest struct {
			Emoji string `json:"emoji" binding:"required"`
		}

		var req EmojiRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
			c.Abort()
			return
		}

		// Basic emoji validation - you can enhance this
		if len(req.Emoji) == 0 || len(req.Emoji) > 10 {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.invalid_emoji")), "messages.invalid_emoji", http.StatusBadRequest)
			c.Abort()
			return
		}

		c.Set("emoji", req.Emoji)
		c.Next()
	})
}
