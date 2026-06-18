package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
)

type AssessmentResource interface {
	FormatAssessment(assessment *models.Assessment) *prot.Assessment
	FormatAssessments(assessments []*models.Assessment) []*prot.Assessment
	FormatModelAssessment(req *prot.AssessmentRequest) *models.Assessment
}

type AssessmentResourceImpl struct{}

func NewAssessmentResource() AssessmentResource {
	return &AssessmentResourceImpl{}
}

func (r *AssessmentResourceImpl) FormatAssessment(assessment *models.Assessment) *prot.Assessment {
	if assessment == nil {
		return nil
	}

	fileInfos := make([]*prot.AssessmentFileInfo, 0, len(assessment.FileInfos))
	files := make([]string, 0, len(assessment.FileInfos))
	for _, info := range assessment.FileInfos {
		url := utils.StaticURL(info.Path, models.Storage)
		fileInfos = append(fileInfos, &prot.AssessmentFileInfo{
			Id:   info.Id,
			Disk: info.Disk,
			Path: info.Path,
		})
		if url != "" {
			files = append(files, url)
		}
	}

	criteriaIDs := assessment.AssessmentCriteriaIds
	if criteriaIDs == nil {
		criteriaIDs = []int64{}
	}

	return &prot.Assessment{
		Id:                        assessment.ID,
		Name:                      assessment.Name,
		Description:               assessment.Description,
		Type:                      assessment.Type,
		ProgramId:                 assessment.ProgramId,
		SubjectId:                 assessment.SubjectId,
		CreatedAt:                 assessment.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:                 assessment.CreatedBy,
		UpdatedAt:                 assessment.UpdatedAt.Format("2006-01-02 15:04:05"),
		UpdatedBy:                 assessment.UpdatedBy,
		AssessmentCriteriaGroupId: assessment.AssessmentCriteriaGroupId,
		StudyReportCriteriaId:     assessment.StudyReportCriteriaId,
		AssessmentCriteriaIds:     criteriaIDs,
		FileInfos:                 fileInfos,
		Files:                     files,
	}
}

func (r *AssessmentResourceImpl) FormatAssessments(assessments []*models.Assessment) []*prot.Assessment {
	result := make([]*prot.Assessment, 0, len(assessments))
	for _, item := range assessments {
		if formatted := r.FormatAssessment(item); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *AssessmentResourceImpl) FormatModelAssessment(req *prot.AssessmentRequest) *models.Assessment {
	if req == nil {
		return nil
	}

	assessmentType := req.Type
	if assessmentType == "" {
		assessmentType = "mini_test"
	}

	mediaRepo := repositories.NewMediaRepository()
	fileInfos := make(models.MediaInfos, 0, len(req.FileInfos))
	for _, file := range req.FileInfos {
		if file == nil {
			continue
		}
		path := utils.StripDomain(file.Path, models.Storage)
		if path == "" {
			continue
		}
		disk := file.Disk
		if disk == "" {
			info := mediaRepo.GetMediaInfo(path, models.Storage)
			disk = info.Disk
		}
		fileInfos = append(fileInfos, models.MediaDetail{
			Id:   file.Id,
			Path: path,
			Disk: disk,
		})
	}

	if len(fileInfos) == 0 {
		for _, fileURL := range req.Files {
			path := utils.StripDomain(fileURL, models.Storage)
			if path == "" {
				continue
			}
			info := mediaRepo.GetMediaInfo(path, models.Storage)
			fileInfos = append(fileInfos, models.MediaDetail{
				Id:   info.Id,
				Path: path,
				Disk: info.Disk,
			})
		}
	}

	criteriaIDs := models.AssessmentCriteriaIDs(req.AssessmentCriteriaIds)
	if criteriaIDs == nil {
		criteriaIDs = []int64{}
	}

	return &models.Assessment{
		ID:                        req.Id,
		Name:                      req.Name,
		Description:               req.Description,
		Type:                      assessmentType,
		ProgramId:                 req.ProgramId,
		SubjectId:                 req.SubjectId,
		AssessmentCriteriaGroupId: req.AssessmentCriteriaGroupId,
		StudyReportCriteriaId:     req.StudyReportCriteriaId,
		AssessmentCriteriaIds:     criteriaIDs,
		FileInfos:                 fileInfos,
	}
}
