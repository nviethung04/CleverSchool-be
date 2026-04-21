package controllers

import (
	"be-lms/services"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestResultController struct {
	service services.ContestResultService
}

func NewContestResultController() *ContestResultController {
	return &ContestResultController{
		service: services.NewContestResultService(),
	}
}

// GetContestRoundAnswers handles GET /api/manage/contest-rounds/:id/answers
func (crc *ContestResultController) GetContestRoundAnswers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	userId, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_user_id")
		return
	}

	answers, err := crc.service.GetContestRoundAnswers(c, int64(id), int64(userId))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	utils.Respond(c, answers, nil, "")
}

// GetContestRoundAnswersByStudent handles GET /api/manage/contest-rounds/:id/answers/student
func (crc *ContestResultController) GetContestRoundAnswersByStudent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	answers, err := crc.service.GetContestRoundAnswersByStudent(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	utils.Respond(c, answers, nil, "")
}

// GetContestRoundResults handles GET /api/manage/contest-rounds/:id/results
func (crc *ContestResultController) GetContestRoundResults(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	results, err := crc.service.GetContestRoundResults(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	utils.Respond(c, results, nil, "")
}

// GetContestRoundLeaderboard handles GET /api/manage/contest-rounds/:id/leaderboard
func (crc *ContestResultController) GetContestRoundLeaderboard(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	leaderboard, err := crc.service.GetContestRoundLeaderboard(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	utils.Respond(c, leaderboard, nil, "")
}
