package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type AssessmentStudentController struct {
	svc services.AssessmentStudentService
}

func NewAssessmentStudentController() *AssessmentStudentController {
	return &AssessmentStudentController{
		svc: services.NewAssessmentStudentService(),
	}
}

// GetStudentsWithAssessment
// GET /api/dashboard/teacher/assessment/students?course_id=&assessment_id=&page=&limit=
func (ctl *AssessmentStudentController) GetStudentsWithAssessment(c *gin.Context) {
	var req requests.AssessmentStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetStudentsWithAssessment(c, &req)
	utils.Respond(c, resp, err, "")
}
