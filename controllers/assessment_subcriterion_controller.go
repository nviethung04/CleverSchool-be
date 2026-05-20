package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type AssessmentSubcriterionController struct {
	*GenericController[models.AssessmentSubcriterion, prot.AssessmentSubcriterion, *prot.AssessmentSubcriterionRequest]
}

func NewAssessmentSubcriterionController(service services.AssessmentSubcriterionService) *AssessmentSubcriterionController {
	resource := resources.NewAssessmentSubcriterionResource()
	adapter := NewAssessmentSubcriterionResourceAdapter(resource)

	genericController := NewGenericController(
		service,
		adapter,
		func() *prot.AssessmentSubcriterionRequest {
			return &prot.AssessmentSubcriterionRequest{}
		},
		func(items []*prot.AssessmentSubcriterion, totalCount uint64) interface{} {
			return &prot.AssessmentSubcriterionListResponse{
				Subcriteria: items,
				Total:       int64(totalCount),
			}
		},
	)

	return &AssessmentSubcriterionController{
		GenericController: genericController,
	}
}

