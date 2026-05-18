package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
)

type TeachingPlanResource interface {
	FormatTeachingPlan(TeachingPlan *models.TeachingPlan) *prot.TeachingPlan
	FormatTeachingPlans(TeachingPlans []*models.TeachingPlan) []*prot.TeachingPlan
	FormatModelTeachingPlan(TeachingPlan *prot.TeachingPlanRequest) *models.TeachingPlan
}

type TeachingPlanResourceImpl struct{}

func NewTeachingPlanResource() TeachingPlanResource {
	return &TeachingPlanResourceImpl{}
}

func (r *TeachingPlanResourceImpl) FormatTeachingPlan(teachingPlan *models.TeachingPlan) *prot.TeachingPlan {
	if teachingPlan == nil {
		return nil
	}

	var approvedBy prot.TeachingPlanApproved
	if teachingPlan.Status {
		approvedBy = prot.TeachingPlanApproved{
			ApprovedBy:   teachingPlan.ApprovedBy,
			ApprovedAt:   teachingPlan.ApprovedAt.Format("2006-01-02 15:04:05"),
			ApprovedNote: teachingPlan.ApprovedNote,
			Status:       teachingPlan.Status,
		}
	}

	var lesson *prot.TeachingPlanLessonInfo
	if len(teachingPlan.Lessons) > 0 {
		lesson = &prot.TeachingPlanLessonInfo{
			Id:          teachingPlan.Lessons[0].ID,
			Title:       teachingPlan.Lessons[0].Title,
			Description: teachingPlan.Lessons[0].Description,
		}
	}

	return &prot.TeachingPlan{
		Id:          int64(teachingPlan.ID),
		Title:       teachingPlan.Title,
		Description: teachingPlan.Description,
		FileUrl:     utils.StaticURL(teachingPlan.FileInfo.Path, models.Storage),
		FileType:    teachingPlan.FileType,
		Approved:    &approvedBy,
		Lesson:      lesson,
		CreatedAt:   teachingPlan.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   teachingPlan.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *TeachingPlanResourceImpl) FormatTeachingPlans(teachingPlans []*models.TeachingPlan) []*prot.TeachingPlan {
	result := make([]*prot.TeachingPlan, 0, len(teachingPlans))
	for _, t := range teachingPlans {
		if formatted := r.FormatTeachingPlan(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *TeachingPlanResourceImpl) FormatModelTeachingPlan(teachingPlan *prot.TeachingPlanRequest) *models.TeachingPlan {
	if teachingPlan == nil {
		return nil
	}

	fileUrl := utils.StripDomain(teachingPlan.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()

	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	return &models.TeachingPlan{
		ID:          int64(teachingPlan.Id),
		Title:       teachingPlan.Title,
		Description: teachingPlan.Description,
		FileInfo:    fileInfo,
		FileType:    teachingPlan.FileType,
	}
}
