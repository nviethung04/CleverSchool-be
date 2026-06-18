package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type AssessmentController struct {
	*GenericController[models.Assessment, prot.Assessment, *prot.AssessmentRequest]
}

func NewAssessmentController(service services.AssessmentService) *AssessmentController {
	assessmentResource := resources.NewAssessmentResource()
	assessmentResourceAdapter := NewAssessmentResourceAdapter(assessmentResource)

	genericController := NewGenericController(
		service,
		assessmentResourceAdapter,
		func() *prot.AssessmentRequest {
			return &prot.AssessmentRequest{}
		},
		func(assessments []*prot.Assessment, totalCount uint64) interface{} {
			return &prot.AssessmentsResponse{
				Assessments: assessments,
				Total:       int64(totalCount),
			}
		},
	)

	return &AssessmentController{
		GenericController: genericController,
	}
}
