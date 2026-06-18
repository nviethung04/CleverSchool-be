package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"net/http"

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
				TotalCount: int64(totalCount),
			}
		},
	)

	return &SettingController{
		GenericController: genericController,
		service:           service,
	}
}

func (ctl *SettingController) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Respond(c, nil, nil, "messages.data_invalid", http.StatusBadRequest)
		return
	}

	setting, err := ctl.service.GetByKey(key)
	if err != nil {
		utils.Respond(c, nil, err, "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, setting, nil, "")
}
