package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppConfigController struct {
	service services.AppConfigService
}

func NewAppConfigController(service services.AppConfigService) *AppConfigController {
	return &AppConfigController{
		service: service,
	}
}

// GetConfig godoc
// @Summary Get app configuration
// @Description Get app configuration based on platform and version from headers
// @Tags Config
// @Accept json
// @Produce json
// @Param platform header string false "Platform (ios, android, web)"
// @Param version header string false "App version (e.g., 1.0.0)"
// @Success 200 {object} prot.AppConfigResponse
// @Router /api/config [get]
func (ctrl *AppConfigController) GetConfig(c *gin.Context) {
	config, err := ctrl.service.GetConfig(c)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data", http.StatusInternalServerError)
		return
	}

	response := &prot.AppConfigResponse{
		AppConfig: config,
	}

	utils.Respond(c, response, nil, "")
}

