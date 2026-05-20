package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
)

type GradeResource interface {
	FormatGrade(grade *models.Grade) *prot.Grade
	FormatGrades(grades []*models.Grade) []*prot.Grade
	FormatModelGrade(grade *prot.GradeRequest) *models.Grade
}

type GradeResourceImpl struct{}

func NewGradeResource() GradeResource {
	return &GradeResourceImpl{}
}

func (r *GradeResourceImpl) FormatGrade(grade *models.Grade) *prot.Grade {
	if grade == nil {
		return nil
	}

	return &prot.Grade{
		Id:        grade.ID,
		Number:    grade.Number,
		NameVn:    grade.NameVN,
		NameEn:    grade.NameEN,
		CreatedAt: grade.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: grade.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *GradeResourceImpl) FormatGrades(grades []*models.Grade) []*prot.Grade {
	result := make([]*prot.Grade, 0, len(grades))
	for _, t := range grades {
		if formatted := r.FormatGrade(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *GradeResourceImpl) FormatModelGrade(grade *prot.GradeRequest) *models.Grade {
	if grade == nil {
		return nil
	}

	return &models.Grade{
		ID:     grade.Id,
		Number: grade.Number,
		NameVN: grade.NameVn,
		NameEN: grade.NameEn,
	}
}

