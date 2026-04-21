package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type AssessmentCriteriaGroupResource interface {
	FormatGroups(items []*models.AssessmentCriteriaGroup) []*prot.AssessmentCriteriaGroup
	FormatGroup(item *models.AssessmentCriteriaGroup) *prot.AssessmentCriteriaGroup
	FormatModelGroup(req *prot.AssessmentCriteriaGroupRequest) *models.AssessmentCriteriaGroup
}

type assessmentCriteriaGroupResource struct {
	criterionResource AssessmentCriterionResource
}

func NewAssessmentCriteriaGroupResource() AssessmentCriteriaGroupResource {
	return &assessmentCriteriaGroupResource{
		criterionResource: NewAssessmentCriterionResource(),
	}
}

func (r *assessmentCriteriaGroupResource) FormatGroups(items []*models.AssessmentCriteriaGroup) []*prot.AssessmentCriteriaGroup {
	result := make([]*prot.AssessmentCriteriaGroup, 0, len(items))
	for _, item := range items {
		if formatted := r.FormatGroup(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *assessmentCriteriaGroupResource) FormatGroup(item *models.AssessmentCriteriaGroup) *prot.AssessmentCriteriaGroup {
	if item == nil {
		return nil
	}

	// Format criteria từ RefCriteria
	criteria := make([]*prot.AssessmentCriterion, 0, len(item.RefCriteria))
	for _, ref := range item.RefCriteria {
		if ref.AssessmentCriterion.ID > 0 {
			criterion := r.criterionResource.FormatAssessmentCriterion(&ref.AssessmentCriterion)
			if criterion != nil {
				criteria = append(criteria, criterion)
			}
		}
	}

	return &prot.AssessmentCriteriaGroup{
		Id:        item.ID,
		Name:      item.Name,
		SubjectId: item.SubjectID,
		CreatedAt: item.CreatedAt.Unix(),
		CreatedBy: item.CreatedBy,
		UpdatedAt: item.UpdatedAt.Unix(),
		UpdatedBy: item.UpdatedBy,
		HasFile:   item.HasFile,
		Criteria:  criteria,
	}
}

func (r *assessmentCriteriaGroupResource) FormatModelGroup(req *prot.AssessmentCriteriaGroupRequest) *models.AssessmentCriteriaGroup {
	if req == nil {
		return nil
	}

	return &models.AssessmentCriteriaGroup{
		ID:        req.Id,
		Name:      req.Name,
		SubjectID: req.SubjectId,
		HasFile:   req.HasFile,
	}
}
