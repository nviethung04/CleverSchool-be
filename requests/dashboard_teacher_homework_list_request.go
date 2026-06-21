package requests

type DashboardTeacherHomeworkListRequest struct {
	CourseID    int64  `form:"course_id" binding:"required"`
	StudentID   int64  `form:"student_id" binding:"required"`
	StartDate   int64  `form:"start_date"`
	EndDate     int64  `form:"end_date"`
	ChapterID   int64  `form:"chapter_id"`
	LessonIDs   string `form:"lesson_ids"`
	HomeworkIDs string `form:"homework_ids"`
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	OrderBy     string `form:"order_by"`
}
