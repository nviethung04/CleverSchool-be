package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
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

