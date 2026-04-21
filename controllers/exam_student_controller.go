package controllers

import (
	"be-lms/prot"
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type ExamStudentController struct {
	service services.ExamStudentService
}

func NewExamStudentController(service services.ExamStudentService) *ExamStudentController {
	return &ExamStudentController{service: service}
}

func (ctrl *ExamStudentController) GetExamStudents(c *gin.Context) {
	var req requests.ExamStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid params"})
		return
	}
	if req.ExamID == 0 {
		c.JSON(400, gin.H{"error": "exam_id required"})
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}
	resp, err := ctrl.service.GetExamStudentsByExamIDService(req.ExamID, req.CourseID, req.Limit, req.Page)
	utils.Respond(c, resp, err, "")
}

func (ctrl *ExamStudentController) GetExamByStudent(c *gin.Context) {
	var req requests.GetExamByStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid params"})
		return
	}

	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	if userID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	req.UserID = userID

	resp, err := ctrl.service.GetExamByStudentService(c, req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if req.CourseID > 0 {
		var course *prot.GetExamByStudentCourse
		if len(resp.Courses) > 0 {
			course = resp.Courses[0]
		}
		utils.Respond(c, course, err, "")
		return
	}
	utils.Respond(c, resp, err, "")
}
