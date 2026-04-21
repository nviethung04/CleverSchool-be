package controllers

import (
	"be-lms/services"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestScoreController struct {
	service services.ContestScoreService
}

func NewContestScoreController() *ContestScoreController {
	return &ContestScoreController{
		service: services.NewContestScoreService(),
	}
}

// SaveContestScoreMultipleChoice handles POST /api/study/contest-score/multiple-choice
func (csc *ContestScoreController) SaveContestScoreMultipleChoice(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreMultipleChoice(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreFillInBlank handles POST /api/study/contest-score/fill-in-blank
func (csc *ContestScoreController) SaveContestScoreFillInBlank(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreFillInBlank(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreOrderingDragdrop handles POST /api/study/contest-score/ordering-and-dragdrop
func (csc *ContestScoreController) SaveContestScoreOrderingDragdrop(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreOrderingDragdrop(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreMatching handles POST /api/study/contest-score/matching
func (csc *ContestScoreController) SaveContestScoreMatching(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreMatching(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreLabeling handles POST /api/study/contest-score/labeling
func (csc *ContestScoreController) SaveContestScoreLabeling(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreLabeling(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreCategory handles POST /api/study/contest-score/category
func (csc *ContestScoreController) SaveContestScoreCategory(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreCategory(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SaveContestScoreManualScoring handles POST /api/study/contest-score/manual-scoring
func (csc *ContestScoreController) SaveContestScoreManualScoring(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SaveContestScoreManualScoring(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_save_score")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SubmitContestRound handles POST /api/study/contest-score/submit-contest-round
func (csc *ContestScoreController) SubmitContestRound(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SubmitContestRound(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_submit_contest")
		return
	}

	utils.Respond(c, result, nil, "")
}

// SkipContestQuestion handles POST /api/study/contest-score/skip-question
func (csc *ContestScoreController) SkipContestQuestion(c *gin.Context) {
	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	result, err := csc.service.SkipContestQuestion(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_skip_question")
		return
	}

	utils.Respond(c, result, nil, "")
}

// CheckSubmitContestRound handles GET /api/study/contest-score/check-submit-contest-round
func (csc *ContestScoreController) CheckSubmitContestRound(c *gin.Context) {
	contestRoundId, err := strconv.Atoi(c.Query("contest_round_id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	result, err := csc.service.CheckSubmitContestRound(c, int64(contestRoundId))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_check_submit")
		return
	}

	utils.Respond(c, result, nil, "")
}
