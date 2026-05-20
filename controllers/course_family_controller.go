package controllers

import (
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"github.com/gin-gonic/gin"
)

type CourseFamilyController struct {
	svc services.CourseFamilyService
}

func NewCourseFamilyController() *CourseFamilyController {
	return &CourseFamilyController{
		svc: services.NewCourseFamilyService(repositories.NewCourseFamilyRepository()),
	}
}

func (ctl *CourseFamilyController) GetCourseFamily(c *gin.Context) {
	var req requests.CourseFamilyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	family, err := ctl.svc.GetCourseFamily(req.CourseID)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	utils.Respond(c, resources.CourseFamilyResource(family), nil, "")
}
