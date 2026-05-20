package controllers

import (
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestStudentController struct {
	service services.ContestStudentService
}

func NewContestStudentController(service services.ContestStudentService) *ContestStudentController {
	return &ContestStudentController{service: service}
}

func (ctrl *ContestStudentController) GetContestRoundsByStudent(c *gin.Context) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	if userID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Get query params
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	resp, err := ctrl.service.GetContestRoundsByStudent(userID, limit, page)
	utils.Respond(c, resp, err, "")
}
