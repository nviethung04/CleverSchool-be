package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type SubjectController struct {
	*GenericController[models.Subject, prot.Subject, *prot.SubjectRequest]
}

func NewSubjectController(service services.SubjectService) *SubjectController {
	subjectResource := resources.NewSubjectResource()
	subjectResourceAdapter := NewSubjectResourceAdapter(subjectResource)

	genericController := NewGenericController(
		service,
		subjectResourceAdapter,
		func() *prot.SubjectRequest {
			return &prot.SubjectRequest{}
		},
		func(subjects []*prot.Subject, totalCount uint64) interface{} {
			return &prot.SubjectsResponse{
				Subjects:       subjects,
				TotalCount: totalCount,
			}
		},
	)

	return &SubjectController{
		GenericController: genericController,
	}
}
