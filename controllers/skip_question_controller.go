package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SkipQuestionController struct {
	service services.SkipQuestionService
}

func NewSkipQuestionController() *SkipQuestionController {
	return &SkipQuestionController{
		service: services.NewSkipQuestionService(),
	}
}

func (c *SkipQuestionController) SkipQuestion(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SkipQuestionRequest](ctx, func() *prot.SkipQuestionRequest {
		return &prot.SkipQuestionRequest{}
	})

	if err != nil {
		utils.Respond(ctx, nil, err, message, http.StatusBadRequest)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	response, err := c.service.SkipQuestion(req, userID)
	utils.Respond(ctx, response, err, "")
}

func (c *SkipQuestionController) CheckSubmitHomework(ctx *gin.Context) {
	homeworkIDStr := ctx.Query("homework_id")
	lessonIDStr := ctx.Query("lesson_id")

	var homeworkID, lessonID int64
	if homeworkIDStr != "" {
		if v, err := strconv.ParseInt(homeworkIDStr, 10, 64); err == nil {
			homeworkID = v
		} else {
			// Nếu parse lỗi, trả về proto mặc định thay vì 400/null
			resp := &prot.CheckSubmitHomeworkResponse{HomeworkId: 0, UserId: 0, SkippedCount: 0, SkippedQuestionIds: []int64{}}
			utils.Respond(ctx, resp, nil, "")
			return
		}
	} else {
		// Nếu thiếu homework_id, trả về proto mặc định thay vì 400/null
		resp := &prot.CheckSubmitHomeworkResponse{HomeworkId: 0, UserId: 0, SkippedCount: 0, SkippedQuestionIds: []int64{}}
		utils.Respond(ctx, resp, nil, "")
		return
	}

	if lessonIDStr != "" {
		if v, err := strconv.ParseInt(lessonIDStr, 10, 64); err == nil {
			lessonID = v
		} else {
			// Nếu parse lỗi, trả về proto mặc định thay vì 400/null
			resp := &prot.CheckSubmitHomeworkResponse{HomeworkId: 0, UserId: 0, SkippedCount: 0, SkippedQuestionIds: []int64{}}
			utils.Respond(ctx, resp, nil, "")
			return
		}
	} else {
		// Nếu thiếu lesson_id, trả về proto mặc định thay vì 400/null
		resp := &prot.CheckSubmitHomeworkResponse{HomeworkId: 0, UserId: 0, SkippedCount: 0, SkippedQuestionIds: []int64{}}
		utils.Respond(ctx, resp, nil, "")
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	// Kiểm tra bộ 3 (homework_id, user_id, lesson_id) từ homework_users
	homeworkUserRepo := repositories.NewHomeworkUserRepository()
	_, err = homeworkUserRepo.GetByHomeworkUserAndLesson(homeworkID, userID, lessonID)
	if err != nil {
		// Trả về lỗi 500 với message "You have not checked the answer"
		utils.Respond(ctx, nil, fmt.Errorf("Bạn chưa nhấn \"Check\" để kiểm tra câu trả lời"), "Bạn chưa nhấn \"Check\" để kiểm tra câu trả lời", http.StatusInternalServerError)
		return
	}

	resp, err := c.service.CheckSubmitHomework(homeworkID, userID)
	utils.Respond(ctx, resp, err, "")
}

