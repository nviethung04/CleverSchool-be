package controllers

import (
	"be-lms/i18n"
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ProvinceController struct {
	svc services.ProvinceService
}

func NewProvinceController(svc services.ProvinceService) *ProvinceController {
	return &ProvinceController{svc: svc}
}

func (ctl *ProvinceController) GetAll(c *gin.Context) {
	provinces, err := ctl.svc.GetAll()
	utils.Respond(c, &prot.ProvincesResponse{
		Provinces:  provinces,
		TotalCount: uint64(len(provinces)),
	}, err, "")
}

func (ctl *ProvinceController) GetByID(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.missing_code")), "messages.missing_code")
		return
	}

	province, err := ctl.svc.GetByCode(code)
	utils.Respond(c, &prot.ProvinceResponse{Province: province}, err, "")
}

func (ctl *ProvinceController) Wards(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.missing_code")), "messages.missing_code")
		return
	}

	wards, err := ctl.svc.Wards(code)
	count := len(wards)

	utils.Respond(c, &prot.WardsResponse{Wards: wards, TotalCount: uint64(count)}, err, "")
}
