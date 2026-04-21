package requests

type DashboardCourseStudentsRequest struct {
	CourseID  int64  `form:"course_id" json:"course_id" binding:"required"`
	StartDate string `form:"start_date" json:"start_date" binding:"required"` // Format: YYYY-MM-DD hoặc Unix timestamp
	EndDate   string `form:"end_date" json:"end_date" binding:"required"`     // Format: YYYY-MM-DD hoặc Unix timestamp
}

