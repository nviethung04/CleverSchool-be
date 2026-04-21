package requests

type DashboardExamRankingRequest struct {
	CourseID  int64  `form:"course_id" binding:"required"`
	LessonID  int64  `form:"lesson_id"`
	ExamID    int64  `form:"exam_id"`
	StartDate *int64 `form:"start_date"`
	EndDate   *int64 `form:"end_date"`
	Limit     int    `form:"limit"`
	Page      int    `form:"page"`
	OrderBy   string `form:"order_by"` // "asc" or "desc"
} 