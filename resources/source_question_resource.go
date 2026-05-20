package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
)

type SourceQuestionResource interface {
	FormatSourceQuestion(sourceQuestion *models.SourceQuestion) *prot.SourceQuestion
	FormatSourceQuestions(sourceQuestions []*models.SourceQuestion) []*prot.SourceQuestion
	FormatModelSourceQuestion(sourceQuestion *prot.SourceQuestionRequest) *models.SourceQuestion
}

type SourceQuestionResourceImpl struct{}

func NewSourceQuestionResource() SourceQuestionResource {
	return &SourceQuestionResourceImpl{}
}

func (r *SourceQuestionResourceImpl) FormatSourceQuestion(sourceQuestion *models.SourceQuestion) *prot.SourceQuestion {
	if sourceQuestion == nil {
		return nil
	}

	var medias []*prot.MediaSourceQuestion

	for _, media := range sourceQuestion.FileInfos {
		medias = append(medias, &prot.MediaSourceQuestion{
			Type: media.Type,
			Url:  utils.StaticURL(media.Path, models.Storage),
		})
	}

	return &prot.SourceQuestion{
		Id:        sourceQuestion.ID,
		Title:     sourceQuestion.Title,
		Skill:     sourceQuestion.Skill,
		Content:   sourceQuestion.Content,
		Level:     sourceQuestion.Level,
		Status:    sourceQuestion.Status,
		Medias:    medias,
		CreatedAt: sourceQuestion.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: sourceQuestion.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SourceQuestionResourceImpl) FormatSourceQuestions(sourceQuestions []*models.SourceQuestion) []*prot.SourceQuestion {
	result := make([]*prot.SourceQuestion, 0, len(sourceQuestions))
	for _, s := range sourceQuestions {
		if formatted := r.FormatSourceQuestion(s); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SourceQuestionResourceImpl) FormatModelSourceQuestion(sourceQuestion *prot.SourceQuestionRequest) *models.SourceQuestion {
	if sourceQuestion == nil {
		return nil
	}

	var fileInfos []models.MediaDetail
	mediaRepo := repositories.NewMediaRepository()

	for _, media := range sourceQuestion.Medias {
		fileUrl := utils.StripDomain(media.Url, models.Storage)
		info := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		fileInfos = append(fileInfos, models.MediaDetail{
			Path: fileUrl,
			Type: media.Type,
			Disk: info.Disk,
			Id:   info.Id,
		})
	}

	return &models.SourceQuestion{
		ID:        int64(sourceQuestion.Id),
		Title:     sourceQuestion.Title,
		Skill:     sourceQuestion.Skill,
		Content:   sourceQuestion.Content,
		Level:     sourceQuestion.Level,
		Status:    sourceQuestion.Status,
		FileInfos: fileInfos,
	}
}
