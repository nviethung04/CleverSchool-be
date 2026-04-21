package resources

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/utils"
)

type FeedbackResource interface {
	FormatFeedback(feedback *dto.FeedbackResponse) *prot.Feedback
	FormatFeedbacks(feedbacks []*dto.FeedbackResponse) []*prot.Feedback
}

type FeedbackResourceImpl struct{}

func NewFeedbackResource() FeedbackResource {
	return &FeedbackResourceImpl{}
}

func (r *FeedbackResourceImpl) FormatFeedback(feedback *dto.FeedbackResponse) *prot.Feedback {
	if feedback == nil {
		return nil
	}

	return &prot.Feedback{
		Id:        feedback.ID,
		UserId:    utils.Int64OrZero(feedback.UserID),
		Content:   feedback.Content,
		Title:     feedback.Title,
		FileUrl:  utils.StaticURL(feedback.FileInfo.Path, models.Storage),
		RoleId:    utils.Int64OrZero(feedback.RoleID),
		Status:    utils.Int64OrZero(feedback.Status),
		Response:  feedback.Response,
		Note:      feedback.Note,
		Type:      utils.Int64OrZero(feedback.Type),
		RoleName:  feedback.RoleName,
		Username:  feedback.Username,
		Name:      feedback.Name,
		CreatedAt: feedback.CreatedAt.Unix(),
		CreatedBy: feedback.CreatedBy,
		UpdatedAt: feedback.UpdatedAt.Unix(),
		UpdatedBy: feedback.UpdatedBy,
	}
}

func (r *FeedbackResourceImpl) FormatFeedbacks(feedbacks []*dto.FeedbackResponse) []*prot.Feedback {
	result := make([]*prot.Feedback, 0, len(feedbacks))
	for _, e := range feedbacks {
		if formatted := r.FormatFeedback(e); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}
