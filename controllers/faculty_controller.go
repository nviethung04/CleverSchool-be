package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/dto"
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

type FacultyController struct {
	*GenericController[models.Faculty, prot.Faculty, *prot.FacultyRequest]
	service services.FacultyService
}

func NewFacultyController(service services.FacultyService) *FacultyController {
	facultyResource := resources.NewFacultyResource()
	facultyResourceAdapter := NewFacultyResourceAdapter(facultyResource)

	genericController := NewGenericController(
		service,
		facultyResourceAdapter,
		func() *prot.FacultyRequest {
			return &prot.FacultyRequest{}
		},
		func(faculties []*prot.Faculty, totalCount uint64) interface{} {
			return &prot.FacultiesResponse{
				Faculties:  faculties,
				TotalCount: totalCount,
			}
		},
	)

	return &FacultyController{
		GenericController: genericController,
		service:           service,
	}
}

func (fc *FacultyController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := fc.service.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (fc *FacultyController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := fc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("Faculty import failed", "error", err)
		} else {
			config.Log.Info("Faculty import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}

