package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type StudyShiftRepository interface {
	base.BaseRepositoryInterface[models.StudyShift]
}

type studyShiftRepository struct {
	*base.BaseRepository[models.StudyShift]
}

func NewStudyShiftRepository() StudyShiftRepository {
	return &studyShiftRepository{
		BaseRepository: base.NewBaseRepository[models.StudyShift](),
	}
}
