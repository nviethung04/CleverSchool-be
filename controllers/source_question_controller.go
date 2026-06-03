package controllers

import (
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SourceQuestionController struct {
	service services.SourceQuestionService
}

func NewSourceQuestionController(service services.SourceQuestionService) *SourceQuestionController {
	return &SourceQuestionController{service}
}

func (sc *SourceQuestionController) GetAll(c *gin.Context) {
	sourceQuestions, totalCount, err := sc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	var sourceQuestionPtrs []*models.SourceQuestion
	for i := range sourceQuestions {
		sourceQuestionPtrs = append(sourceQuestionPtrs, &sourceQuestions[i])
	}

	sourceQuestionResource := resources.NewSourceQuestionResource()
	sourceQuestionsResponse := sourceQuestionResource.FormatSourceQuestions(sourceQuestionPtrs)

	list := &prot.SourceQuestionsResponse{
		SourceQuestions: sourceQuestionsResponse,
		TotalCount:      uint64(totalCount),
	}

	utils.Respond(c, list, err, "")
}

func (sc *SourceQuestionController) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	sourceQuestion, err := sc.service.GetByID(c, id)
	if err != nil {
		utils.Respond(c, nil, err,  "messages.data_existed", http.StatusNotFound)
		return
	}

	sourceQuestionResource := resources.NewSourceQuestionResource()
	formattedSourceQuestion := sourceQuestionResource.FormatSourceQuestion(sourceQuestion)

	utils.Respond(c, formattedSourceQuestion, err, "")
}

func (sc *SourceQuestionController) Create(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.SourceQuestionRequest](c, func() *prot.SourceQuestionRequest {
		return &prot.SourceQuestionRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	sourceQuestion, err := sc.service.Create(c, req)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	sourceQuestionResource := resources.NewSourceQuestionResource()
	formattedSourceQuestion := sourceQuestionResource.FormatSourceQuestion(sourceQuestion)

	utils.Respond(c, formattedSourceQuestion, err, "")
}

func (sc *SourceQuestionController) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	req, err, message := utils.GetBody[*prot.SourceQuestionRequest](c, func() *prot.SourceQuestionRequest {
		return &prot.SourceQuestionRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	req.Id = int64(id)

	sourceQuestion, err := sc.service.Update(c, req)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	sourceQuestionResource := resources.NewSourceQuestionResource()
	formattedSourceQuestion := sourceQuestionResource.FormatSourceQuestion(sourceQuestion)

	utils.Respond(c, formattedSourceQuestion, err, "")

	utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_record_update")), "messages.no_record_update")
}

func (sc *SourceQuestionController) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := sc.service.Delete(c, id)
	utils.Respond(c, &prot.DeleteResponse{
		Id: int64(id),
	}, err, "")
}

func (sc *SourceQuestionController) Restore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	source, err := sc.service.Restore(c, id)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	sourceQuestionResource := resources.NewSourceQuestionResource()
	formattedSourceQuestion := sourceQuestionResource.FormatSourceQuestion(source)

	utils.Respond(c, formattedSourceQuestion, nil, "")
}
