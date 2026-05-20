package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExerciseCommentController struct {
	service services.ExerciseCommentService
}

func NewExerciseCommentController(service services.ExerciseCommentService) *ExerciseCommentController {
	return &ExerciseCommentController{service: service}
}

func (ctl *ExerciseCommentController) PostExerciseComment(c *gin.Context) {
	var req struct {
		ExerciseID int64  `json:"exercise_id" binding:"required"`
		StudentID  int64  `json:"student_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, &prot.ExamCommentResponse{Success: false, Message: err.Error()}, err, "messages.input_invalid", http.StatusBadRequest)
		return
	}
	tokenStr := c.GetHeader("Token")
	teacherID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, &prot.ExamCommentResponse{Success: false, Message: "Invalid token"}, err, "Invalid token", http.StatusUnauthorized)
		return
	}
	resp := ctl.service.CreateExerciseComment(req.ExerciseID, req.StudentID, teacherID, req.Content)
	utils.Respond(c, resp, nil, "")
}

