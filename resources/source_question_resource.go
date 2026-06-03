package resources

import (
	"be-lms/models"
	"be-lms/prot"
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

	return &prot.SourceQuestion{
		Id:        int64(sourceQuestion.ID),
		Title:     sourceQuestion.Title,
		Skill:     sourceQuestion.Skill,
		Content:   sourceQuestion.Content,
		Level:     sourceQuestion.Level,
		Status:    sourceQuestion.Status,
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

	return &models.SourceQuestion{
		ID:      int64(sourceQuestion.Id),
		Title:   sourceQuestion.Title,
		Skill:   sourceQuestion.Skill,
		Content: sourceQuestion.Content,
		Level:   sourceQuestion.Level,
		Status:  sourceQuestion.Status,
	}
}
