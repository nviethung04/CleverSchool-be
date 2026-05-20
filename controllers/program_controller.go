package controllers

import (
	"fmt"

	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProgramController struct {
	*GenericController[models.Program, prot.Program, *prot.ProgramRequest]
	service services.ProgramService
}

func NewProgramController(service services.ProgramService) *ProgramController {
	programResource := resources.NewProgramResource()
	programResourceAdapter := NewProgramResourceAdapter(programResource)

	genericController := NewGenericController(
		service,
		programResourceAdapter,
		func() *prot.ProgramRequest {
			return &prot.ProgramRequest{}
		},
		func(programs []*prot.Program, totalCount uint64) interface{} {
			return &prot.ProgramsResponse{
				Programs:   programs,
				TotalCount: totalCount,
			}
		},
	)

	return &ProgramController{
		GenericController: genericController,
		service:           service,
	}
}

func (pc *ProgramController) SortChapters(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	req, err, message := utils.GetBody[*prot.ProgramChapterSort](c, func() *prot.ProgramChapterSort {
		return &prot.ProgramChapterSort{}
	})

	err = pc.service.SortChapters(c, id, req)

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	utils.Respond(c, req, err, "")
}

func (pc *ProgramController) Export(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func(programID int) {
		url, serviceErr := pc.service.Export(cCp, programID)
		resultChan <- &dto.MyExportResult{Url: url, Err: serviceErr}
	}(id)

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (pc *ProgramController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		if importErr := pc.service.Import(cCp, file); importErr != nil {
			config.Log.Error("Program import failed", "error", importErr)
		} else {
			config.Log.Info("Program import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "messages.import_complete")
}
