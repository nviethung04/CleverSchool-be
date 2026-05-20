package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetExamAnswersController struct {
	service services.GetExamAnswersService
}

func NewGetExamAnswersController(service services.GetExamAnswersService) *GetExamAnswersController {
	return &GetExamAnswersController{
		service: service,
	}
}

func (c *GetExamAnswersController) GetTeacherExamAnswers(ctx *gin.Context) {
	// Lấy exam_id và user_id từ query string
	examIDStr := ctx.Query("exam_id")
	userIDStr := ctx.Query("user_id")

	// Validate exam_id
	if examIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "exam_id is required",
		})
		return
	}

	// Validate user_id
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	// Parse exam_id
	examID, err := strconv.ParseInt(examIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid exam_id",
		})
		return
	}

	// Parse user_id
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id",
		})
		return
	}

	// Tạo request
	req := &prot.GetExamAnswersRequest{
		ExamId: examID,
		UserId: userID,
	}

	// Gọi service
	response, err := c.service.GetExamAnswers(ctx, req)
	utils.Respond(ctx, response, err, "")
}

func (c *GetExamAnswersController) GetStudentExamAnswers(ctx *gin.Context) {
	// Lấy exam_id và user_id từ query string
	examIDStr := ctx.Query("exam_id")

	// Validate exam_id
	if examIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "exam_id is required",
		})
		return
	}

	// Parse exam_id
	examID, err := strconv.ParseInt(examIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid exam_id",
		})
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	// Tạo request
	req := &prot.GetExamAnswersRequest{
		ExamId: examID,
		UserId: userID,
	}

	// Gọi service
	response, err := c.service.GetExamAnswers(ctx, req)
	utils.Respond(ctx, response, err, "")
}

