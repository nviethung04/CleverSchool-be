package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type TrainingLevelController struct {
	*GenericController[models.TrainingLevel, prot.TrainingLevel, *prot.TrainingLevelRequest]
}

func NewTrainingLevelController(service services.TrainingLevelService) *TrainingLevelController {
	trainingLevelResource := resources.NewTrainingLevelResource()
	trainingLevelResourceAdapter := NewTrainingLevelResourceAdapter(trainingLevelResource)

	genericController := NewGenericController(
		service,
		trainingLevelResourceAdapter,
		func() *prot.TrainingLevelRequest {
			return &prot.TrainingLevelRequest{}
		},
		func(trainingLevels []*prot.TrainingLevel, totalCount uint64) interface{} {
			return &prot.TrainingLevelsResponse{
				TrainingLevels: trainingLevels,
				TotalCount:     totalCount,
			}
		},
	)

	return &TrainingLevelController{
		GenericController: genericController,
	}
}

