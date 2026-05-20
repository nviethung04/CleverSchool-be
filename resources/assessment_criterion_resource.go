package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
)

type AssessmentCriterionResource interface {
	FormatAssessmentCriterion(item *models.AssessmentCriterion) *prot.AssessmentCriterion
	FormatAssessmentCriteria(items []*models.AssessmentCriterion) []*prot.AssessmentCriterion
	FormatModelAssessmentCriterion(req *prot.AssessmentCriterionRequest) *models.AssessmentCriterion
}

type assessmentCriterionResourceImpl struct{}

func NewAssessmentCriterionResource() AssessmentCriterionResource {
	return &assessmentCriterionResourceImpl{}
}

func (r *assessmentCriterionResourceImpl) FormatAssessmentCriterion(item *models.AssessmentCriterion) *prot.AssessmentCriterion {
	if item == nil {
		return nil
	}

	return &prot.AssessmentCriterion{
		Id:             item.ID,
		SubjectId:      item.SubjectID,
		Name:           item.Name,
		Description:    item.Description,
		HasSubcriteria: item.HasSubcriteria,
		MaxScore:       item.MaxScore,
		CreatedAt:      item.CreatedAt.Unix(),
		CreatedBy:      item.CreatedBy,
		UpdatedAt:      item.UpdatedAt.Unix(),
		UpdatedBy:      item.UpdatedBy,
		Subcriteria:    r.formatSubcriteria(item.AssessmentSubcriteria),
	}
}

func (r *assessmentCriterionResourceImpl) formatSubcriteria(items []models.AssessmentSubcriterion) []*prot.AssessmentCriterionSubItem {
	result := make([]*prot.AssessmentCriterionSubItem, 0, len(items))
	for _, sub := range items {
		result = append(result, &prot.AssessmentCriterionSubItem{
			Id:          sub.ID,
			Name:        sub.Name,
			Description: sub.Description,
			MaxScore:    sub.MaxScore,
		})
	}
	return result
}

func (r *assessmentCriterionResourceImpl) FormatAssessmentCriteria(items []*models.AssessmentCriterion) []*prot.AssessmentCriterion {
	result := make([]*prot.AssessmentCriterion, 0, len(items))
	for _, item := range items {
		if formatted := r.FormatAssessmentCriterion(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *assessmentCriterionResourceImpl) FormatModelAssessmentCriterion(req *prot.AssessmentCriterionRequest) *models.AssessmentCriterion {
	if req == nil {
		return nil
	}

	return &models.AssessmentCriterion{
		ID:             req.Id,
		SubjectID:      req.SubjectId,
		Name:           req.Name,
		Description:    req.Description,
		HasSubcriteria: req.HasSubcriteria,
		MaxScore:       req.MaxScore,
	}
}
