package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"sort"
)

type LessonPlanPartResource interface {
	FormatLessonPlanPart(lesson *models.LessonPlanPart) *prot.LessonPlanPart
	FormatLessonPlanParts(lessons []*models.LessonPlanPart) []*prot.LessonPlanPart
	FormatModelLessonPlanPart(lessonPlanPart *prot.LessonPlanPartRequest) *models.LessonPlanPart
}

type LessonPlanPartResourceImpl struct{}

func NewLessonPlanPartResource() LessonPlanPartResource {
	return &LessonPlanPartResourceImpl{}
}

func (r *LessonPlanPartResourceImpl) FormatLessonPlanPart(part *models.LessonPlanPart) *prot.LessonPlanPart {
	if part == nil {
		return nil
	}

	isProgram := part.CourseID == 0

	return &prot.LessonPlanPart{
		Id:           part.ID,
		LessonPlanId: part.LessonPlanID,
		CourseId: part.CourseID,
		Title:        part.Title,
		Tag:          part.Tag,
		ObjectTitle:	part.ObjectTitle,
		CoverImage:   utils.StaticURL(part.CoverImageInfo.Path, models.Storage),
		SortPosition: int32(part.SortPosition),
		Time:         part.Time,
		IsClasswork:  part.IsClasswork,
		FileType:     part.FileType,
		Link:         utils.StaticURL(part.LinkInfo.Path, models.Storage),
		LinkType:     part.LinkType,
		GuideTeacher: part.GuideTeacher,
		GuideStudent: part.GuideStudent,
		CreatedAt:    part.CreatedAt.Unix(),
		CreatedBy:    part.CreatedBy,
		UpdatedAt:    part.UpdatedAt.Unix(),
		UpdatedBy:    part.UpdatedBy,
		File:         part.File,
		IsProgram: isProgram,
	}
}

func (r *LessonPlanPartResourceImpl) FormatLessonPlanParts(lessonPlanParts []*models.LessonPlanPart) []*prot.LessonPlanPart {
	result := make([]*prot.LessonPlanPart, 0, len(lessonPlanParts))
	for _, t := range lessonPlanParts {
		if formatted := r.FormatLessonPlanPart(t); formatted != nil {
			result = append(result, formatted)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
        if result[i].IsProgram && !result[j].IsProgram {
            return true
        }
        if !result[i].IsProgram && result[j].IsProgram {
            return false
        }
        return result[i].Id < result[j].Id
    })

	return result
}

func (r *LessonPlanPartResourceImpl) FormatModelLessonPlanPart(lessonPlanPart *prot.LessonPlanPartRequest) *models.LessonPlanPart {
	if lessonPlanPart == nil {
		return nil
	}

	link := utils.StripDomain(lessonPlanPart.Link, models.Storage)
	coverImage := utils.StripDomain(lessonPlanPart.CoverImage, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	linkInfo := mediaRepo.GetMediaInfo(link, models.Storage)
	coverImageInfo := mediaRepo.GetMediaInfo(coverImage, models.Storage)

	return &models.LessonPlanPart{
		ID:             lessonPlanPart.Id,
		Title:          lessonPlanPart.Title,
		LessonPlanID:   lessonPlanPart.LessonPlanId,
		CourseID: lessonPlanPart.CourseId,
		Tag:            lessonPlanPart.Tag,
		ObjectTitle:	lessonPlanPart.ObjectTitle,
		CoverImageInfo: coverImageInfo,
		SortPosition:   int16(lessonPlanPart.SortPosition),
		Time:           lessonPlanPart.Time,
		IsClasswork:    lessonPlanPart.IsClasswork,
		FileType:       lessonPlanPart.FileType,
		LinkInfo:       linkInfo,
		LinkType:       lessonPlanPart.LinkType,
		GuideTeacher:   lessonPlanPart.GuideTeacher,
		GuideStudent:   lessonPlanPart.GuideStudent,
		File:           lessonPlanPart.File,
		MaxScore:       float64(lessonPlanPart.MaxScore),
	}
}
