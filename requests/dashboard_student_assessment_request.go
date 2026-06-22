package requests

type DashboardStudentAssessmentRequest struct {
	CourseID     *int64 `form:"course_id"`
	AssessmentID *int64 `form:"assessment_id"`
	SubjectID    *int64 `form:"subject_id"`
	Type         string `form:"type"`
	Limit        int    `form:"limit"`
	Page         int    `form:"page"`
}
