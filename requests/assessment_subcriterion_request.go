package requests

type GetAssessmentSubcriterionRequest struct {
	Limit       int    `form:"limit"`
	Page        int    `form:"page"`
	Keyword     string `form:"keyword"`
	CriterionID *int64 `form:"criterion_id"`
}
