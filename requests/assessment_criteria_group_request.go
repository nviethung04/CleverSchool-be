package requests

type GetAssessmentCriteriaGroupRequest struct {
	Limit     int    `form:"limit"`
	Page      int    `form:"page"`
	Keyword   string `form:"keyword"`
	Name      string `form:"name"`
	SubjectID int64  `form:"subject_id"`
}
