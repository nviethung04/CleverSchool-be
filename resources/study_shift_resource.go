package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type StudyShiftResource interface {
	FormatStudyShift(studyShift *models.StudyShift) *prot.StudyShift
	FormatStudyShifts(studyShifts []*models.StudyShift) []*prot.StudyShift
	FormatModelStudyShift(studyShift *prot.StudyShiftRequest) *models.StudyShift
}

type StudyShiftResourceImpl struct{}

func NewStudyShiftResource() StudyShiftResource {
	return &StudyShiftResourceImpl{}
}

func (r *StudyShiftResourceImpl) FormatStudyShift(studyShift *models.StudyShift) *prot.StudyShift {
	if studyShift == nil {
		return nil
	}

	return &prot.StudyShift{
		Id:        int64(studyShift.ID),
		Name:      studyShift.Name,
		StartTime: studyShift.StartTime,
		EndTime:   studyShift.EndTime,
		CreatedAt: studyShift.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: studyShift.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *StudyShiftResourceImpl) FormatStudyShifts(studyShifts []*models.StudyShift) []*prot.StudyShift {
	result := make([]*prot.StudyShift, 0, len(studyShifts))
	for _, u := range studyShifts {
		if formatted := r.FormatStudyShift(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *StudyShiftResourceImpl) FormatModelStudyShift(studyShift *prot.StudyShiftRequest) *models.StudyShift {
	if studyShift == nil {
		return nil
	}

	return &models.StudyShift{
		ID:        int64(studyShift.Id),
		Name:      studyShift.Name,
		StartTime: studyShift.StartTime,
		EndTime:   studyShift.EndTime,
	}
}
