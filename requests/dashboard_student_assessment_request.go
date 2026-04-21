package requests

type DashboardStudentAssessmentRequest struct {
	CourseID     int64  `form:"course_id" binding:"required"`
	AssessmentID int64  `form:"assessment_id"`
	Type         string `form:"type"` // Filter theo type: 'mini_test' hoặc 'final_exam'
}

