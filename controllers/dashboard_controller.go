package controllers

import (
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	service services.DashboardService
}

func NewDashboardController(service services.DashboardService) *DashboardController {
	return &DashboardController{service}
}

func (dc *DashboardController) Dashboard(c *gin.Context) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId == models.AdminRoleId || roleId == models.SchoolRoleId {
		dashboard, err := dc.service.DashboardAdmin(c)
		if err != nil {
			utils.Respond(c, nil, err, "")
			return
		}

		utils.Respond(c, dashboard, err, "")
		return
	}

	utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.role_invalid")), "messages.role_invalid")
}

func (dc *DashboardController) Export(c *gin.Context) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId == models.AdminRoleId || roleId == models.SchoolRoleId {

		cCp := c.Copy()
		resultChan := make(chan *dto.MyExportResult)

		go func() {
			url, err := dc.service.Export(cCp)
			resultChan <- &dto.MyExportResult{Url: url, Err: err}
		}()

		res := <-resultChan
		utils.Respond(c, gin.H{"url": res.Url}, res.Err, "")
		return
	}

	utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.role_invalid")), "messages.role_invalid")
}
