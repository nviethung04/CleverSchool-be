package controllers

import (
	"be-lms/config"
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SchoolController struct {
	*GenericController[models.School, prot.School, *prot.SchoolRequest]
	service services.SchoolService
}

func NewSchoolController(service services.SchoolService) *SchoolController {
	schoolResource := resources.NewSchoolResource()
	schoolResourceAdapter := NewSchoolResourceAdapter(schoolResource)

	genericController := NewGenericController(
		service,
		schoolResourceAdapter,
		func() *prot.SchoolRequest {
			return &prot.SchoolRequest{}
		},
		func(schools []*prot.School, totalCount uint64) interface{} {
			return &prot.SchoolListResponse{
				Schools:    schools,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &SchoolController{
		GenericController: genericController,
		service:           service,
	}
}

func (sc *SchoolController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := sc.service.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (sc *SchoolController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := sc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("User import failed", "error", err)
		} else {
			config.Log.Info("User import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "messages.import_complete")
}

func (sc *SchoolController) GetSchoolStudents(c *gin.Context) {
	schoolId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	students, err := sc.service.GetSchoolStudents(c, schoolId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	utils.Respond(c, students, nil, "")
}
