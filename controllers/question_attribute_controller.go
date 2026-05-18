package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type QuestionAttributeController struct {
	service services.QuestionAttributeService
	*GenericController[models.QuestionAttribute, prot.QuestionAttribute, *prot.QuestionAttributeRequest]
}

func NewQuestionAttributeController(service services.QuestionAttributeService) *QuestionAttributeController {
	questionAttributeResource := resources.NewQuestionAttributeResource()
	questionAttributeResourceAdapter := NewQuestionAttributeResourceAdapter(questionAttributeResource)

	genericController := NewGenericController(
		service,
		questionAttributeResourceAdapter,
		func() *prot.QuestionAttributeRequest {
			return &prot.QuestionAttributeRequest{}
		},
		func(questionAttributes []*prot.QuestionAttribute, totalCount uint64) interface{} {
			return &prot.QuestionAttributesResponse{
				Attributes:    questionAttributes,
				TotalCount: totalCount,
			}
		},
	)

	return &QuestionAttributeController{
		GenericController: genericController,
		service:           service,
	}
}

func (qc *QuestionAttributeController) GetParents(c *gin.Context) {
	attributes, err := qc.service.GetParents(c)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	var attrPtrs []*models.QuestionAttribute
	for i := range attributes {
		attrPtrs = append(attrPtrs, &attributes[i])
	}

	attributeResource := resources.NewQuestionAttributeResource()
	formattedAttributes := attributeResource.FormatQuestionAttributes(attrPtrs)

	response := &prot.QuestionAttributesResponse{
		Attributes: formattedAttributes,
	}

	utils.Respond(c, response, nil, "")
}
