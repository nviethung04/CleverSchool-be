package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExamController struct {
	svc services.ExamService
	*GenericController[models.Exam, prot.Exam, *prot.ExamRequest]
}

func NewExamController(service services.ExamService) *ExamController {
	examResource := resources.NewExamResource()
	examResourceAdapter := NewExamResourceAdapter(examResource)

	genericController := NewGenericController(
		service,
		examResourceAdapter,
		func() *prot.ExamRequest {
			return &prot.ExamRequest{}
		},
		func(exams []*prot.Exam, totalCount uint64) interface{} {
			return &prot.ExamListResponse{
				Exams:    exams,
				Total: int64(totalCount),
			}
		},
	)

	return &ExamController{
		GenericController: genericController,
		svc:           service,
	}
}

func (ctl *ExamController) Cloned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	exam, err := ctl.svc.Cloned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.data_existed", http.StatusNotFound)
		return
	}

	utils.Respond(c, exam, err, "")
}

func (ctl *ExamController) Assigned(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	assigned, err := ctl.svc.Assigned(c, id)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, assigned, err, "")
}

func (ctl *ExamController) AssignedLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	lessons, err := ctl.svc.AssignedLesson(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, lessons, err, "")
}
