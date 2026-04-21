package controllers

import (
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HomeworkAnswerController struct {
	service services.HomeworkAnswerService
}

func NewHomeworkAnswerController(service services.HomeworkAnswerService) *HomeworkAnswerController {
	return &HomeworkAnswerController{service: service}
}

func (ctl *HomeworkAnswerController) GetStudentHomeworkAnswer(c *gin.Context) {
	homeworkIDStr := c.Query("homework_id")
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid homework_id"})
		return
	}

	lessonIDStr := c.Query("lesson_id")
	var lessonID int64
	if lessonIDStr != "" {
		lessonID, err = strconv.ParseInt(lessonIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson_id"})
			return
		}
	}

	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.service.GetHomeworkAnswerProto(homeworkID, userID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Respond(c, resp, err, "")
}

func (ctl *HomeworkAnswerController) GetTeacherHomeworkAnswer(c *gin.Context) {
	homeworkIDStr := c.Query("homework_id")
	userIDStr := c.Query("user_id")
	lessonIDStr := c.Query("lesson_id")
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid homework_id"})
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	var lessonID int64
	if lessonIDStr != "" {
		lessonID, err = strconv.ParseInt(lessonIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lesson_id"})
			return
		}
	}
	resp, err := ctl.service.GetHomeworkAnswerProto(homeworkID, userID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.Respond(c, resp, err, "")
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
