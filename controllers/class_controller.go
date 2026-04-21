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

type ClassController struct {
	*GenericController[models.Class, prot.ClassResponse, *prot.ClassRequest]
	service services.ClassService
}

func NewClassController(service services.ClassService) *ClassController {
	classResource := resources.NewClassResource()
	classResourceAdapter := NewClassResourceAdapter(classResource)

	genericController := NewGenericController(
		service,
		classResourceAdapter,
		func() *prot.ClassRequest {
			return &prot.ClassRequest{}
		},
		func(classes []*prot.ClassResponse, totalCount uint64) interface{} {
			return &prot.ClassListResponse{
				Classes:    classes,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &ClassController{
		GenericController: genericController,
		service:           service,
	}
}

func (cc *ClassController) GetUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	users, err := cc.service.GetUsers(c, int64(id))
	cc.respondWithUsers(c, users, err)
}

func (cc *ClassController) StoreUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	students, err := cc.service.StoreUsers(c, int64(id))
	cc.respondWithUsers(c, students, err)
}

func (cc *ClassController) AddUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	students, err := cc.service.AddUsers(c, int64(id))
	cc.respondWithUsers(c, students, err)
}

func (cc *ClassController) respondWithUsers(c *gin.Context, students []models.User, err error) {
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	ptrs := make([]*models.User, len(students))
	for i := range students {
		ptrs[i] = &students[i]
	}

	userResource := resources.NewUserResource()
	formattedUsers := userResource.FormatUsers(ptrs)

	utils.Respond(c, formattedUsers, nil, "")
}

func (cc *ClassController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := cc.service.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (cc *ClassController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := cc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("Class import failed", "error", err)
		} else {
			config.Log.Info("Class import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}
