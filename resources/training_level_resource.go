package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type TrainingLevelResource interface {
	FormatTrainingLevel(trainingLevel *models.TrainingLevel) *prot.TrainingLevel
	FormatTrainingLevels(trainingLevels []*models.TrainingLevel) []*prot.TrainingLevel
	FormatModelTrainingLevel(trainingLevel *prot.TrainingLevelRequest) *models.TrainingLevel
}

type TrainingLevelResourceImpl struct{}

func NewTrainingLevelResource() TrainingLevelResource {
	return &TrainingLevelResourceImpl{}
}

func (r *TrainingLevelResourceImpl) FormatTrainingLevel(trainingLevel *models.TrainingLevel) *prot.TrainingLevel {
	if trainingLevel == nil {
		return nil
	}

	return &prot.TrainingLevel{
		Id:           int64(trainingLevel.ID),
		Name:         trainingLevel.Name,
		Description:  trainingLevel.Description,
		SortPosition: trainingLevel.SortPosition,
		CreatedAt:    trainingLevel.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    trainingLevel.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *TrainingLevelResourceImpl) FormatTrainingLevels(trainingLevels []*models.TrainingLevel) []*prot.TrainingLevel {
	result := make([]*prot.TrainingLevel, 0, len(trainingLevels))
	for _, t := range trainingLevels {
		if formatted := r.FormatTrainingLevel(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *TrainingLevelResourceImpl) FormatModelTrainingLevel(trainingLevel *prot.TrainingLevelRequest) *models.TrainingLevel {
	if trainingLevel == nil {
		return nil
	}

	return &models.TrainingLevel{
		ID:           int64(trainingLevel.Id),
		Name:         trainingLevel.Name,
		Description:  trainingLevel.Description,
		SortPosition: trainingLevel.SortPosition,
	}
}
