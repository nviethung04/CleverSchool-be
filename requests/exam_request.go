package requests

type GetExamRequest struct {
	LessonID   *int  `form:"lesson_id"`
	IsAssigned *bool `form:"is_assigned"` // Optional field to filter by assignment status
	Limit      int   `form:"limit"`
	Page       int   `form:"page"`
}
