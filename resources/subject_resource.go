package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type SubjectResource interface {
	FormatSubject(subject *models.Subject) *prot.Subject
	FormatSubjects(subjects []*models.Subject) []*prot.Subject
	FormatModelSubject(subject *prot.SubjectRequest) *models.Subject
}

type SubjectResourceImpl struct{}

func NewSubjectResource() SubjectResource {
	return &SubjectResourceImpl{}
}

func (r *SubjectResourceImpl) FormatSubject(subject *models.Subject) *prot.Subject {
	if subject == nil {
		return nil
	}

	return &prot.Subject{
		Id:          int64(subject.ID),
		Name:        subject.Name,
		Description: subject.Description,
		Status:      subject.Status,
		CreatedAt:   subject.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   subject.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SubjectResourceImpl) FormatSubjects(subjects []*models.Subject) []*prot.Subject {
	result := make([]*prot.Subject, 0, len(subjects))
	for _, s := range subjects {
		if formatted := r.FormatSubject(s); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SubjectResourceImpl) FormatModelSubject(subject *prot.SubjectRequest) *models.Subject {
	if subject == nil {
		return nil
	}

	return &models.Subject{
		ID:          int64(subject.Id),
		Name:        subject.Name,
		Description: subject.Description,
		Status:      subject.Status,
	}
}
