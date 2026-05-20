package controllers

import (
	// "be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"

	// "net/http"
	// "strconv"

	"github.com/gin-gonic/gin"
)

type ContestCourseController struct {
	service services.ContestCourseService
}

func NewContestCourseController(service services.ContestCourseService) *ContestCourseController {
	return &ContestCourseController{service: service}
}

func (ccc *ContestCourseController) GetContestCourseDetail(c *gin.Context) {
	detail, err := ccc.service.GetContestCourseDetail(int64(0), nil)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}
	utils.Respond(c, detail, nil, "")
}

// AssignContestToCourse handles POST /api/manage/contests/:id/assign-course
func (ccc *ContestCourseController) AssignContestToCourse(c *gin.Context) {
	response, err := ccc.service.AssignContestToCourse(int64(0), int64(0))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_assign_contest")
		return
	}
	utils.Respond(c, response, nil, "")
}

// RemoveContestFromCourse handles DELETE /api/manage/contests/:id/remove-course
func (ccc *ContestCourseController) RemoveContestFromCourse(c *gin.Context) {
	response, err := ccc.service.RemoveContestFromCourse(int64(0), int64(0))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_remove_contest")
		return
	}
	utils.Respond(c, response, nil, "")
}

