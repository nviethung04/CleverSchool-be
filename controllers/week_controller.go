package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type WeekController struct {
	service services.WeekService
}

func NewWeekController(service services.WeekService) *WeekController {
	return &WeekController{service}
}

func (cc *WeekController) GetAll(c *gin.Context) {
	weeks, totalCount, err := cc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	var weekPtrs []*models.Week
	for i := range weeks {
		weekPtrs = append(weekPtrs, &weeks[i])
	}

	weekResource := resources.NewWeekResource()
	weeksResponse := weekResource.FormatWeeks(weekPtrs)

	list := &prot.WeeksResponse{
		Weeks:      weeksResponse,
		TotalCount: uint64(totalCount),
	}

	utils.Respond(c, list, err, "")
}

func (wc *WeekController) GetWeekByDate(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		current := time.Now()
		dateStr = current.Format("2006-01-02")
	}

	week, err := wc.service.GetByDate(dateStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	weekResource := resources.NewWeekResource()
	weekResponse := weekResource.FormatWeek(week)

	utils.Respond(c, weekResponse, nil, "")
}

