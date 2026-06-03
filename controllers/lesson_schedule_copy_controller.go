package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type LessonScheduleCopyController struct {
	svc services.LessonScheduleCopyService
}

func NewLessonScheduleCopyController() *LessonScheduleCopyController {
	return &LessonScheduleCopyController{
		svc: services.NewLessonScheduleCopyService(),
	}
}

func (ctl *LessonScheduleCopyController) CopyLessonSchedules(c *gin.Context) {
	var req requests.CopyLessonScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Dữ liệu đầu vào không hợp lệ")
		return
	}

	resp, err := ctl.svc.CopyLessonSchedules(c, &req)
	utils.Respond(c, resp, err, "")
}
