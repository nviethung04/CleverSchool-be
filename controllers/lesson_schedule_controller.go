package controllers

import (
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"

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
