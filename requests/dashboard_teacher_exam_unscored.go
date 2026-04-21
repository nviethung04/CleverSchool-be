package requests

type DashboardTeacherExamUnscoredListRequest struct {
	Page      int               `json:"page" form:"page"`
	Limit     int               `json:"limit" form:"limit"`
	Sort      map[string]string `json:"sort" form:"sort"`
	StartDate int64             `json:"start_date" form:"start_date"`
	EndDate   int64             `json:"end_date" form:"end_date"`
	CourseID  int64             `json:"course_id" form:"course_id"`
	ExamID    int64             `json:"exam_id" form:"exam_id"`
} 