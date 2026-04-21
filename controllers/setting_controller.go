package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type SettingController struct {
	*GenericController[models.Setting, prot.Setting, *prot.SettingRequest]
	service services.SettingService
}

func NewSettingController(service services.SettingService) *SettingController {
	settingResource := resources.NewSettingResource()
	settingResourceAdapter := NewSettingResourceAdapter(settingResource)

	genericController := NewGenericController(
		service,
		settingResourceAdapter,
		func() *prot.SettingRequest {
			return &prot.SettingRequest{}
		},
		func(settings []*prot.Setting, totalCount uint64) interface{} {
			return &prot.SettingsResponse{
				Settings:   settings,
				TotalCount: totalCount,
			}
		},
	)

	ctl := &SettingController{
		GenericController: genericController,
		service:           service,
	}

	ctl.GenericController.WithUsedError(func() bool {
		return true
	})

	return ctl
}

func (s *SettingController) GetByKey(c *gin.Context) {
	key := c.Param("key")

	item, err := s.service.GetByKey(c, key)
	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, item, nil, "")
}
