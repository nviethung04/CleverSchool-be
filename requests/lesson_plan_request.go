package requests

type GetLessonPlanRequest struct {
	LessonID *int   `form:"lesson_id"`
	Limit    int    `form:"limit"`
	Page     int    `form:"page"`
	Keyword  string `form:"keyword"`
}
