package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestQuestionController struct {
	service services.ContestQuestionService
}

func NewContestQuestionController() *ContestQuestionController {
	return &ContestQuestionController{
		service: services.NewContestQuestionService(),
	}
}

// GetContestRoundQuestions handles GET /api/manage/contest-rounds/:id/questions
func (cqc *ContestQuestionController) GetContestRoundQuestions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	questions, err := cqc.service.GetContestRoundQuestions(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	utils.Respond(c, questions, nil, "")
}

// AssignQuestionsToContestRound handles POST /api/manage/contest-rounds/:id/questions
func (cqc *ContestQuestionController) AssignQuestionsToContestRound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	// fmt.Printf("🔍 Controller: Received request for contest_round_id=%d, req=%v\n", id, req)

	// Parse request to get question IDs
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		utils.Respond(c, nil, fmt.Errorf("invalid request format"), "messages.invalid_request")
		return
	}

	questionIdsInterface, exists := reqMap["question_ids"]
	if !exists {
		utils.Respond(c, nil, fmt.Errorf("question_ids not found in request"), "messages.invalid_request")
		return
	}

	questionIds, ok := questionIdsInterface.([]interface{})
	if !ok {
		utils.Respond(c, nil, fmt.Errorf("question_ids must be an array"), "messages.invalid_request")
		return
	}

	// Convert to CreateQuestionRelationRequest format
	questionScores := make(map[string]float64)
	for _, idInterface := range questionIds {
		idFloat, ok := idInterface.(float64)
		if !ok {
			continue // Skip invalid IDs
		}
		questionId := int64(idFloat)
		questionScores[strconv.FormatInt(questionId, 10)] = 1.0 // Default score
	}

	// Create CreateQuestionRelationRequest
	relationReq := &prot.CreateQuestionRelationRequest{
		ContestRoundId: int32(id),
		QuestionScores: questionScores,
		IsReset:        false,
	}

	// Use QuestionRelationService instead of ContestQuestionService
	questionRelationService := services.NewQuestionRelationService(repositories.NewQuestionRelationRepository())
	err = questionRelationService.AssignQuestions(c, relationReq)
	if err != nil {
		// fmt.Printf("🔍 Controller: Error from question relation service: %v\n", err)
		utils.Respond(c, nil, err, "messages.error_assign_questions")
		return
	}

	// fmt.Printf("🔍 Controller: Successfully assigned questions using question-relation API\n")
	utils.Respond(c, map[string]interface{}{
		"message": "Questions assigned successfully",
	}, nil, "")
}

// RemoveQuestionsFromContestRound handles DELETE /api/manage/contest-rounds/:id/questions
func (cqc *ContestQuestionController) RemoveQuestionsFromContestRound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	err = cqc.service.RemoveQuestionsFromContestRound(c, int64(id), req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_remove_questions")
		return
	}

	utils.Respond(c, map[string]interface{}{
		"message": "Questions removed successfully",
	}, nil, "")
}
