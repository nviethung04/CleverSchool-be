package controllers

import (
	"be-cleverschool/prot"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type SchoolDashboardController struct{
	service services.SchoolDashboardService
}

func NewSchoolDashboardController(service services.SchoolDashboardService) *SchoolDashboardController {
	return &SchoolDashboardController{
		service: service,
	}
}

func (ctl *SchoolDashboardController) GetSchoolSummary(c *gin.Context) {
    schools, err := ctl.service.GetAllSchoolsWithStats()
    if err != nil {
        utils.Respond(c, nil, fmt.Errorf("Lỗi truy vấn trường học"), "Lỗi truy vấn trường học")
        return
    }
  	list := &prot.SchoolSummaryList{
		Items: schools,
	}

	utils.Respond(c, list, nil, "")
}

