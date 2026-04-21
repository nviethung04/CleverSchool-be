package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"
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
