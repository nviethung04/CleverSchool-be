package requests

type ExamStudentRequest struct {
	ExamID   int64 `form:"exam_id"`
	CourseID int64 `form:"course_id"`
	Limit    int   `form:"limit"`
	Page     int   `form:"page"`
}

type GetExamByStudentRequest struct {
	WeekID    int64  `form:"week_id"`
	UserID    int64  `form:"user_id"`
	StartDate *int64 `form:"start_date"`
	EndDate   *int64 `form:"end_date"`
	Limit     int    `form:"limit"`
	Page      int    `form:"page"`
	CourseID  int64  `form:"course_id"`
}
