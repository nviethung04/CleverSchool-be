package requests

type DashboardTeacherHomeworkStudentRequest struct {
	CourseID   int64               `form:"course_id" binding:"required"`
	HomeworkID int64               `form:"homework_id"`
	LessonID   int64               `form:"lesson_id"`
	Limit      int                 `form:"limit"`
	Page       int                 `form:"page"`
	Sort       map[string]string   `form:"sort"`
}

type DashboardTeacherHomeworkStudentStatsRequest struct {
	CourseID    int64  `form:"course_id" binding:"required"`
	ChapterID   int64  `form:"chapter_id"`
	LessonID    int64  `form:"lesson_id"`
	LessonIDs   string `form:"lesson_ids"` // Danh sách lesson_id cách nhau bởi dấu phẩy (ví dụ: "1,2,3")
	HomeworkIDs string `form:"homework_ids"` // Danh sách homework_id cách nhau bởi dấu phẩy (ví dụ: "1,2,3")
	Limit       int    `form:"limit"`
	Page        int    `form:"page"`
	Sort        map[string]string `form:"sort"`
	OrderBy     string `form:"order_by"` // "completed_homework" hoặc "average_score"
	StartDate   string `form:"start_date"` // Format: YYYY-MM-DD hoặc Unix timestamp
	EndDate     string `form:"end_date"`   // Format: YYYY-MM-DD hoặc Unix timestamp
}

type DashboardTeacherHomeworkOverviewRequest struct {
	CourseID   int64 `form:"course_id"`
	HomeworkID int64 `form:"homework_id" binding:"required"`
	LessonID   int64 `form:"lesson_id"`
} 