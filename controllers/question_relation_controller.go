package controllers

import (
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type QuestionRelationController struct {
	svc services.QuestionRelationService
}

func NewQuestionRelationController(svc services.QuestionRelationService) *QuestionRelationController {
	return &QuestionRelationController{svc: svc}
}

func (ctl *QuestionRelationController) Create(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.CreateQuestionRelationRequest](c, func() *prot.CreateQuestionRelationRequest {
		return &prot.CreateQuestionRelationRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	err = ctl.svc.AssignQuestions(c, req)

	utils.Respond(c, &prot.QuestionRelationResponse{
		Message: "ok",
	}, err, "")
}
