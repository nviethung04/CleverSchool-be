package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExamCommentController struct {
	service services.ExamCommentService
}

func NewExamCommentController(service services.ExamCommentService) *ExamCommentController {
	return &ExamCommentController{service: service}
}

func (ctl *ExamCommentController) PostExamComment(c *gin.Context) {
    req, err, message := utils.GetBody[*prot.ExamCommentRequest](c, func() *prot.ExamCommentRequest {
		return &prot.ExamCommentRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	tokenStr := c.GetHeader("Token")
	teacherID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, &prot.ExamCommentResponse{Success: false, Message: "Invalid token"}, err, "Invalid token", http.StatusUnauthorized)
		return
	}
	resp := ctl.service.CreateExamComment(req.ExamId, req.StudentId, teacherID, req.Content)
	utils.Respond(c, resp, nil, "")
}

