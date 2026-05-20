package controllers

import (
	"fmt"

	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type LessonScheduleController struct {
	service services.LessonScheduleService
}

func NewLessonScheduleController(service services.LessonScheduleService) *LessonScheduleController {
	return &LessonScheduleController{service}
}

func (cc *LessonScheduleController) GetAll(c *gin.Context) {
	groups, err := cc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	lessonScheduleResource := resources.NewLessonScheduleResource()
	groupedResponse := lessonScheduleResource.FormatScheduleGroups(groups)

	list := &prot.LessonSchedulesResponse{
		Groups: groupedResponse,
	}

	utils.Respond(c, list, nil, "")
}

func (cc *LessonScheduleController) SyncAllProgram(c *gin.Context) {
	var req prot.LessonScheduleSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request body", 400)
		return
	}

	if req.CourseId <= 0 {
		utils.Respond(c, nil, fmt.Errorf("course_id phải lớn hơn 0"), "", 400)
		return
	}

	resp, err := cc.service.SyncAllProgram(req.CourseId)
	if err != nil {
		utils.Respond(c, nil, err, "", 500)
		return
	}

	utils.Respond(c, resp, nil, "")
}