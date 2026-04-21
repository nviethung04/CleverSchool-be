package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/utils"
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

	subjectVTGInfo := &prot.SubjectVTGInfo{
		Code:       "",
		Level:      "",
		Enrollment: "",
		Time:       "",
		Avatar:     "",
	}

	if subject.Detail.VTGData != (models.SubjectVTGData{}) {
		subjectVTGInfo.Code = subject.Detail.VTGData.Code
		subjectVTGInfo.Level = subject.Detail.VTGData.Level
		subjectVTGInfo.Enrollment = subject.Detail.VTGData.Enrollment
		subjectVTGInfo.Time = subject.Detail.VTGData.Time
		if subject.Detail.VTGData.Avatar != "" {
			subjectVTGInfo.Avatar = utils.StaticURL(subject.Detail.VTGData.Avatar, models.Storage)
		}
	}

	var faculty prot.Faculty
	if subject.Faculty.ID > 0 {
		faculty = prot.Faculty{
			Id:          int64(subject.Faculty.ID),
			Name:        subject.Faculty.Name,
			Code:        subject.Faculty.Code,
			Description: subject.Faculty.Description,
		}
	}

	var trainingLevels []*prot.TrainingLevel
	if subject.TrainingLevels != nil {
		for _, trainingLevel := range subject.TrainingLevels {
			trainingLevels = append(trainingLevels, &prot.TrainingLevel{
				Id:          int64(trainingLevel.ID),
				Name:        trainingLevel.Name,
				Description: trainingLevel.Description,
			})
		}
	}

	return &prot.Subject{
		Id:             int64(subject.ID),
		FacultyId:      subject.FacultyId,
		Faculty:        &faculty,
		TrainingLevels: trainingLevels,
		Name:           subject.Name,
		Description:    subject.Description,
		Status:         subject.Status,
		VtgInfo:        subjectVTGInfo,
		CreatedAt:      subject.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      subject.UpdatedAt.Format("2006-01-02 15:04:05"),
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

	var subjectVTGInfo models.SubjectVTGData

	if subject.VtgInfo != nil {
		avatar := ""
		if subject.VtgInfo.Avatar != "" {
			avatar = utils.StripDomain(subject.VtgInfo.Avatar, models.Storage)
		}
		subjectVTGInfo = models.SubjectVTGData{
			Code:       subject.VtgInfo.Code,
			Level:      subject.VtgInfo.Level,
			Enrollment: subject.VtgInfo.Enrollment,
			Time:       subject.VtgInfo.Time,
			Avatar:     avatar,
		}
	}

	subjectDetail := models.SubjectDetail{
		VTGData: subjectVTGInfo,
	}

	return &models.Subject{
		ID:          int64(subject.Id),
		FacultyId:   int64(subject.FacultyId),
		Name:        subject.Name,
		Description: subject.Description,
		Status:      subject.Status,
		Detail:      subjectDetail,
	}
}
