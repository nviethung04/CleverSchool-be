package requests

type DashboardTeacherHomeworkScoredListRequest struct {
	Page        int               `json:"page" form:"page"`
	Limit       int               `json:"limit" form:"limit"`
	Sort        map[string]string `json:"sort" form:"sort"`
	StartDate   int64             `json:"start_date" form:"start_date"`
	EndDate     int64             `json:"end_date" form:"end_date"`
	CourseID    int64             `json:"course_id" form:"course_id"`
	HomeworkID  int64             `json:"homework_id" form:"homework_id"`
	UserID      int64             `json:"user_id" form:"user_id"`
	StudentName string            `json:"student_name" form:"student_name"`
}
