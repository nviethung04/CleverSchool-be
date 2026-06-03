package controllers

import (
	_ "be-lms/prot"
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type DashboardListEntityController struct {
	svc services.DashboardListEntityService
}

func NewDashboardListEntityController() *DashboardListEntityController {
	return &DashboardListEntityController{
		svc: services.NewDashboardListEntityService(),
	}
}

func (ctl *DashboardListEntityController) GetSchools(c *gin.Context) {
	var req requests.DashboardSchoolListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetSchools(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetCourses(c *gin.Context) {
	var req requests.DashboardCourseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetCourses(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetTeachers(c *gin.Context) {
	var req requests.DashboardTeacherListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetTeachers(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetSubjects(c *gin.Context) {
	var req requests.DashboardSubjectListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetSubjects(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetExams(c *gin.Context) {
	var req requests.DashboardExamListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetExams(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetHomeworks(c *gin.Context) {
	var req requests.DashboardHomeworkListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetHomeworks(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardListEntityController) GetLessons(c *gin.Context) {
	var req requests.DashboardLessonListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetLessons(c, &req)
	utils.Respond(c, resp, err, "")
}
