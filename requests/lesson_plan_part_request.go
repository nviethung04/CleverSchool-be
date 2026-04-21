package requests

type GetLessonPlanPartRequest struct {
	LessonPlanID int `form:"lesson_plan_id"`
	CourseID int `form:"course_id"`
	Limit        int `form:"limit"`
	Page         int `form:"page"`
}
