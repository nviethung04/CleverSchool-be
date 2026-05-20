package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
)

type QuestionAttributeResource interface {
	FormatQuestionAttribute(attr *models.QuestionAttribute) *prot.QuestionAttribute
	FormatQuestionAttributes(attrs []*models.QuestionAttribute) []*prot.QuestionAttribute
	FormatModelQuestionAttribute(attr *prot.QuestionAttributeRequest) *models.QuestionAttribute
}

type questionAttributeResource struct{}

func NewQuestionAttributeResource() QuestionAttributeResource {
	return &questionAttributeResource{}
}

func (r *questionAttributeResource) FormatQuestionAttribute(attr *models.QuestionAttribute) *prot.QuestionAttribute {
	if attr == nil {
		return nil
	}

	var parentId, subjectId int64
	if attr.ParentID != nil {
		parentId = *attr.ParentID
	}
	if attr.SubjectId != nil {
		subjectId = *attr.SubjectId
	}

	return &prot.QuestionAttribute{
		Id:          attr.ID,
		ParentId:    parentId,
		SubjectId:   subjectId,
		Name:        attr.Name,
		Level:       int32(attr.Level),
		Weight:      attr.Weight,
		Description: attr.Description,
		Nodes:       r.formatNodes(attr.Nodes),
		Status:      attr.Status,
	}
}

func (r *questionAttributeResource) FormatQuestionAttributes(attrs []*models.QuestionAttribute) []*prot.QuestionAttribute {
	result := make([]*prot.QuestionAttribute, 0, len(attrs))
	for _, attr := range attrs {
		if formatted := r.FormatQuestionAttribute(attr); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *questionAttributeResource) formatNodes(nodes []models.QuestionAttribute) []*prot.QuestionAttribute {
	var ptrs []*models.QuestionAttribute
	for i := range nodes {
		ptrs = append(ptrs, &nodes[i])
	}
	return r.FormatQuestionAttributes(ptrs)
}

func (r *questionAttributeResource) FormatModelQuestionAttribute(req *prot.QuestionAttributeRequest) *models.QuestionAttribute {
	if req == nil {
		return nil
	}

	var parentId, subjectId *int64
	if req.ParentId != 0 {
		parentId = &req.ParentId
	}
	if req.SubjectId != 0 {
		subjectId = &req.SubjectId
	}

	return &models.QuestionAttribute{
		ID:          req.Id,
		ParentID:    parentId,
		SubjectId:   subjectId,
		Name:        req.Name,
		Level:       int(req.Level),
		Weight:      req.Weight,
		Description: req.Description,
		Status:      req.Status,
	}
}

