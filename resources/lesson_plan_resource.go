package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
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

	var author prot.LessonPlanAuthor
	if lp.Author != nil {
		author = prot.LessonPlanAuthor{
			Id:     int64(lp.Author.ID),
			Name:   lp.Author.Name,
			Avatar: utils.StaticURL(lp.Author.AvatarInfo.Path, models.Storage),
		}
	}

	var lessons []*prot.LessonPlanLessonInfo
	if lp.Lessons != nil {
		for _, lesson := range lp.Lessons {
			lessonInfo := &prot.LessonPlanLessonInfo{
				Id:          int64(lesson.ID),
				Title:       lesson.Title,
				Description: lesson.Description,
			}

			if lesson.Chapter.ID > 0 {
				chapter := &prot.LessonPlanChapterInfo{
					Id:          int64(lesson.Chapter.ID),
					Title:       lesson.Chapter.Title,
					Description: lesson.Chapter.Description,
				}

				if lesson.Chapter.Program.ID > 0 {
					program := &prot.LessonPlanProgramInfo{
						Id:          int64(lesson.Chapter.Program.ID),
						Name:        lesson.Chapter.Program.Name,
						Description: lesson.Chapter.Program.Description,
					}

					if len(lesson.Chapter.Program.Subjects) > 0 {
						for _, subject := range lesson.Chapter.Program.Subjects {
							subjectInfo := &prot.LessonPlanSubjectInfo{
								Id:          int64(subject.ID),
								Name:        subject.Name,
								Description: subject.Description,
							}

							if subject.Faculty.ID > 0 {
								subjectInfo.Faculty = &prot.LessonPlanFacultyInfo{
									Id:          int64(subject.Faculty.ID),
									Name:        subject.Faculty.Name,
									Code:        subject.Faculty.Code,
									Description: subject.Faculty.Description,
								}
							}

							program.Subjects = append(program.Subjects, subjectInfo)
						}
					}

					chapter.Program = program
				}

				lessonInfo.Chapter = chapter
			}
			lessons = append(lessons, lessonInfo)
		}
	}

	return &prot.LessonPlan{
		Id:           int64(lp.ID),
		Name:         lp.Name,
		Description:  lp.Description,
		CoverImage:   utils.StaticURL(lp.CoverImageInfo.Path, models.Storage),
		ObjectTitle:  lp.ObjectTitle,
		Status:       int32(lp.Status),
		SortPosition: int32(lp.SortPosition),
		TotalTime:    int64(lp.TotalTime),
		Views:        int32(lp.Views),
		CreatedAt:    lp.CreatedAt.Unix(),
		UpdatedAt:    lp.UpdatedAt.Unix(),
		IsComplete:   isComplete,
		CompleteAt:   completeAt,
		AuthorId:     lp.AuthorId,
		Author:       &author,
		Lessons:      lessons,
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
		ObjectTitle:    lessonPlan.ObjectTitle,
		SortPosition:   int(lessonPlan.SortPosition),
		Status:         int(lessonPlan.Status),
		TotalTime:      int(lessonPlan.TotalTime),
		Views:          int(lessonPlan.Views),
	}
}

