package requests

type GetAssessmentCriterionRequest struct {
	Limit          int    `form:"limit"`
	Page           int    `form:"page"`
	Keyword        string `form:"keyword"`
	SubjectID      *int64 `form:"subject_id"`
	HasSubcriteria *bool  `form:"has_subcriteria"`
}
