package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
)

type LessonPlanPartController struct {
	*GenericController[models.LessonPlanPart, prot.LessonPlanPart, *prot.LessonPlanPartRequest]
	svc services.LessonPlanPartService
}

func NewLessonPlanPartController(service services.LessonPlanPartService) *LessonPlanPartController {
	lessonPlanPartResource := resources.NewLessonPlanPartResource()
	lessonPlanPartResourceAdapter := NewLessonPlanPartResourceAdapter(lessonPlanPartResource)

	genericController := NewGenericController(
		service,
		lessonPlanPartResourceAdapter,
		func() *prot.LessonPlanPartRequest {
			return &prot.LessonPlanPartRequest{}
		},
		func(lessonPlanParts []*prot.LessonPlanPart, totalCount uint64) interface{} {
			return &prot.LessonPlanPartListResponse{
				Parts: lessonPlanParts,
				Total:       int64(totalCount),
			}
		},
	)

	return &LessonPlanPartController{
		GenericController: genericController,
		svc:               service,
	}
}
