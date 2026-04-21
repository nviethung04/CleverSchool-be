package requests

type GetHomeworkRequest struct {
	LessonID   *int   `form:"lesson_id"`
	IsAssigned *bool  `form:"is_assigned"`
	Limit      int    `form:"limit"`
	Page       int    `form:"page"`
	Keyword    string `form:"keyword"`
}
