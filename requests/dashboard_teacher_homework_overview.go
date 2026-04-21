package requests

type DashboardTeacherHomeworkOverviewStatsRequest struct {
	StartDate   int64 `json:"start_date" form:"start_date"`
	EndDate     int64 `json:"end_date" form:"end_date"`
	CourseID    int64 `json:"course_id" form:"course_id"`
	LessonID    int64 `json:"lesson_id" form:"lesson_id"`
	HomeworkID  int64 `json:"homework_id" form:"homework_id"`
	UserID      int64 `json:"user_id" form:"user_id"`
}