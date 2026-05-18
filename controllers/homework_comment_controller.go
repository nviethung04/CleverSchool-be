package controllers

import (
    "be-Clever School/prot"
    "be-Clever School/services"
    "be-Clever School/utils"
    "net/http"

    "github.com/gin-gonic/gin"
)

type HomeworkCommentController struct {
    service services.HomeworkCommentService
}

func NewHomeworkCommentController(service services.HomeworkCommentService) *HomeworkCommentController {
    return &HomeworkCommentController{service: service}
}

func (ctl *HomeworkCommentController) PostHomeworkComment(c *gin.Context) {
    var req struct {
        HomeworkID int64  `json:"homework_id" binding:"required"`
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
    resp := ctl.service.CreateHomeworkComment(req.HomeworkID, req.StudentID, teacherID, req.Content)
    utils.Respond(c, resp, nil, "")
}


