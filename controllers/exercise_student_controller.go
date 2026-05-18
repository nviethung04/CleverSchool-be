package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type ExerciseStudentController struct {
    service services.ExerciseStudentService
}

func NewExerciseStudentController(service services.ExerciseStudentService) *ExerciseStudentController {
    return &ExerciseStudentController{service: service}
}

func (ctrl *ExerciseStudentController) GetExerciseStudents(c *gin.Context) {
    var req requests.ExerciseStudentRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid params"})
        return
    }
    if req.ExerciseID == 0 {
        c.JSON(400, gin.H{"error": "exercise_id required"})
        return
    }
    if req.Limit == 0 { req.Limit = 20 }
    if req.Page == 0 { req.Page = 1 }

    resp, err := ctrl.service.GetExerciseStudentsByExerciseIDService(req.ExerciseID, req.CourseID, req.Limit, req.Page)
    utils.Respond(c, resp, err, "")
}


