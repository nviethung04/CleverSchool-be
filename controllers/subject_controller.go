package controllers

import (
	"be-lms/config"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

type SubjectController struct {
	*GenericController[models.Subject, prot.Subject, *prot.SubjectRequest]
	service services.SubjectService
}

func NewSubjectController(service services.SubjectService) *SubjectController {
	subjectResource := resources.NewSubjectResource()
	subjectResourceAdapter := NewSubjectResourceAdapter(subjectResource)

	genericController := NewGenericController(
		service,
		subjectResourceAdapter,
		func() *prot.SubjectRequest {
			return &prot.SubjectRequest{}
		},
		func(subjects []*prot.Subject, totalCount uint64) interface{} {
			return &prot.SubjectsResponse{
				Subjects:       subjects,
				TotalCount: totalCount,
			}
		},
	)

	return &SubjectController{
		GenericController: genericController,
		service:           service,
	}
}

func (sc *SubjectController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *prot.Export, 1)

	go func() {
		url, err := sc.service.Export(cCp)
		if err != nil {
			utils.Respond(cCp, nil, err, "")
			resultChan <- nil
			return
		}
		resultChan <- &prot.Export{Url: url}
	}()

	res := <-resultChan
	if res == nil {
		return
	}
	utils.Respond(c, res, nil, "")
}

func (sc *SubjectController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := sc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("Subject import failed", "error", err)
		} else {
			config.Log.Info("Subject import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}
