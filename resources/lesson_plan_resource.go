package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
)

type LessonPlanResource interface {
	FormatLessonPlan(lesson *models.LessonPlan) *prot.LessonPlan
	FormatLessonPlans(lessons []*models.LessonPlan) []*prot.LessonPlan
	FormatModelLessonPlan(lessonPlan *prot.LessonPlanRequest) *models.LessonPlan
}

type LessonPlanResourceImpl struct{}

func NewLessonPlanResource() LessonPlanResource {
	return &LessonPlanResourceImpl{}
}

func (r *LessonPlanResourceImpl) FormatLessonPlan(lp *models.LessonPlan) *prot.LessonPlan {
	if lp == nil {
		return nil
	}

	var isComplete bool
	var completeAt int64

	if lp.Complete != nil {
		isComplete = true
		completeAt = lp.Complete.CompletedAt.Unix()
	}

	return &prot.LessonPlan{
		Id:           int64(lp.ID),
		Name:         lp.Name,
		Description:  lp.Description,
		CoverImage:   utils.StaticURL(lp.CoverImageInfo.Path, models.Storage),
		ObjectTitle:	 lp.ObjectTitle,
		Status:       int32(lp.Status),
		SortPosition: int32(lp.SortPosition),
		TotalTime:    int64(lp.TotalTime),
		Views:        int32(lp.Views),
		CreatedAt:    lp.CreatedAt.Unix(),
		UpdatedAt:    lp.UpdatedAt.Unix(),
		IsComplete:   isComplete,
		CompleteAt:   completeAt,
	}
}

func (r *LessonPlanResourceImpl) FormatLessonPlans(lessonPlans []*models.LessonPlan) []*prot.LessonPlan {
	result := make([]*prot.LessonPlan, 0, len(lessonPlans))
	for _, t := range lessonPlans {
		if formatted := r.FormatLessonPlan(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *LessonPlanResourceImpl) FormatModelLessonPlan(lessonPlan *prot.LessonPlanRequest) *models.LessonPlan {
	if lessonPlan == nil {
		return nil
	}

	coverImageUrl := utils.StripDomain(lessonPlan.CoverImage, models.Storage)
	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(coverImageUrl, models.Storage)

	return &models.LessonPlan{
		ID:             lessonPlan.Id,
		Name:           lessonPlan.Name,
		Description:    lessonPlan.Description,
		CoverImageInfo: imageInfo,
		ObjectTitle:	lessonPlan.ObjectTitle,
		SortPosition:   int(lessonPlan.SortPosition),
		Status:         int(lessonPlan.Status),
		TotalTime:      int(lessonPlan.TotalTime),
		Views:          int(lessonPlan.Views),
	}
}
