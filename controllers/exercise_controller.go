package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExerciseController struct {
    svc services.ExerciseService
    *GenericController[models.Exercise, prot.Exercise, *prot.ExerciseRequest]
}

func NewExerciseController(service services.ExerciseService) *ExerciseController {
    res := resources.NewExerciseResource()
    adapter := NewExerciseResourceAdapter(res)

    genericController := NewGenericController(
        service,
        adapter,
        func() *prot.ExerciseRequest {
            return &prot.ExerciseRequest{}
        },
        func(items []*prot.Exercise, totalCount uint64) interface{} {
            return &prot.ExerciseListResponse{
                Exercises:    items,
                Total: int64(totalCount),
            }
        },
    )

    return &ExerciseController{
        GenericController: genericController,
        svc:           service,
    }
}

func (ctl *ExerciseController) Cloned(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    exercise, err := ctl.svc.Cloned(c, id)
    if err != nil {
        utils.Respond(c, nil, err, "messages.data_existed", http.StatusNotFound)
        return
    }

    utils.Respond(c, exercise, err, "")
}

func (ctl *ExerciseController) Assigned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	assigned, err := ctl.svc.Assigned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, assigned, err, "")
}

func (ctl *ExerciseController) AssignedLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	lessons, err := ctl.svc.AssignedLesson(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, lessons, err, "")
}
