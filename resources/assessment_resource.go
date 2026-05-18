package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/utils"
)

type AssessmentResource interface {
	FormatAssessment(item *models.Assessment) *prot.Assessment
	FormatAssessments(items []*models.Assessment) []*prot.Assessment
	FormatModelAssessment(req *prot.AssessmentRequest) *models.Assessment
}

type assessmentResourceImpl struct {
	criterionResource AssessmentCriterionResource
}

func NewAssessmentResource() AssessmentResource {
	return &assessmentResourceImpl{
		criterionResource: NewAssessmentCriterionResource(),
	}
}

func (r *assessmentResourceImpl) FormatAssessment(item *models.Assessment) *prot.Assessment {
	if item == nil {
		return nil
	}

	result := &prot.Assessment{
		Id:                          item.ID,
		Name:                        item.Name,
		Description:                 item.Description,
		Type:                        item.Type,
		ProgramId:                   item.ProgramID,
		CreatedAt:                   item.CreatedAt.Unix(),
		CreatedBy:                   item.CreatedBy,
		UpdatedAt:                   item.UpdatedAt.Unix(),
		UpdatedBy:                   item.UpdatedBy,
		SubjectId:                   item.SubjectID,
		StudyReportCriteriaId:       item.StudyReportCriteriaID,
		AssessmentCriteriaGroupId:   item.AssessmentCriteriaGroupID,
		AssessmentCriteriaGroupName: item.AssessmentCriteriaGroupName,
		SubjectName:                 item.SubjectName,
		StudyReportCriteriaName:     item.StudyReportCriteriaName,
		LessonId:                    item.LessonID,
		LessonTitle:                 item.LessonTitle,
		IsAssigned:                  item.IsAssigned,
		HasFile:                     item.AssessmentCriteriaGroupHasFile, // Lấy từ assessment_criteria_groups
		PublishCourseIds:            item.PublishCourseIds,
	}

	// Map criteria từ AssessmentRefCriteria
	if len(item.AssessmentRefCriteria) > 0 {
		criteria := make([]*prot.AssessmentCriterion, 0, len(item.AssessmentRefCriteria))
		for _, ref := range item.AssessmentRefCriteria {
			if ref.AssessmentCriterion.ID > 0 {
				criterion := r.criterionResource.FormatAssessmentCriterion(&ref.AssessmentCriterion)
				if criterion != nil {
					criteria = append(criteria, criterion)
				}
			}
		}
		result.Criteria = criteria
	}

	for _, info := range item.FileInfos {
		url := utils.StaticURL(info.Path, info.Disk)
		result.Files = append(result.Files, url)
	}

	return result
}

func (r *assessmentResourceImpl) FormatAssessments(items []*models.Assessment) []*prot.Assessment {
	result := make([]*prot.Assessment, 0, len(items))
	for _, item := range items {
		if formatted := r.FormatAssessment(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *assessmentResourceImpl) FormatModelAssessment(req *prot.AssessmentRequest) *models.Assessment {
	if req == nil {
		return nil
	}

	entity := &models.Assessment{
		ID:                    req.Id,
		Name:                  req.Name,
		Description:           req.Description,
		Type:                  req.Type,
		ProgramID:             req.ProgramId,
		SubjectID:             req.SubjectId,
		StudyReportCriteriaID: req.StudyReportCriteriaId,
		HasFile:               req.HasFile,
	}

	if len(req.FileInfos) > 0 {
		infos := make(models.MediaInfos, 0, len(req.FileInfos))
		for _, info := range req.FileInfos {
			infos = append(infos, models.MediaDetail{
				Id:   info.Id,
				Disk: info.Disk,
				Path: info.Path,
			})
		}
		entity.FileInfos = infos
	}

	return entity
}
