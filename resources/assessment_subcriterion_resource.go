package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type AssessmentSubcriterionResource interface {
	FormatAssessmentSubcriterion(item *models.AssessmentSubcriterion) *prot.AssessmentSubcriterion
	FormatAssessmentSubcriteria(items []*models.AssessmentSubcriterion) []*prot.AssessmentSubcriterion
	FormatModelAssessmentSubcriterion(req *prot.AssessmentSubcriterionRequest) *models.AssessmentSubcriterion
}

type assessmentSubcriterionResourceImpl struct{}

func NewAssessmentSubcriterionResource() AssessmentSubcriterionResource {
	return &assessmentSubcriterionResourceImpl{}
}

func (r *assessmentSubcriterionResourceImpl) FormatAssessmentSubcriterion(item *models.AssessmentSubcriterion) *prot.AssessmentSubcriterion {
	if item == nil {
		return nil
	}

	return &prot.AssessmentSubcriterion{
		Id:          item.ID,
		CriterionId: item.CriterionID,
		Name:        item.Name,
		Description: item.Description,
		MaxScore:    item.MaxScore,
		CreatedAt:   item.CreatedAt.Unix(),
		CreatedBy:   item.CreatedBy,
		UpdatedAt:   item.UpdatedAt.Unix(),
		UpdatedBy:   item.UpdatedBy,
	}
}

func (r *assessmentSubcriterionResourceImpl) FormatAssessmentSubcriteria(items []*models.AssessmentSubcriterion) []*prot.AssessmentSubcriterion {
	result := make([]*prot.AssessmentSubcriterion, 0, len(items))
	for _, item := range items {
		if formatted := r.FormatAssessmentSubcriterion(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *assessmentSubcriterionResourceImpl) FormatModelAssessmentSubcriterion(req *prot.AssessmentSubcriterionRequest) *models.AssessmentSubcriterion {
	if req == nil {
		return nil
	}

	return &models.AssessmentSubcriterion{
		ID:          req.Id,
		CriterionID: req.CriterionId,
		Name:        req.Name,
		Description: req.Description,
		MaxScore:    req.MaxScore,
	}
}
