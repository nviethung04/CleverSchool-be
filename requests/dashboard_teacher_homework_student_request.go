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
	CourseID  int64 `form:"course_id" binding:"required"`
	LessonID  int64 `form:"lesson_id"`
	Limit     int   `form:"limit"`
	Page      int   `form:"page"`
	Sort      map[string]string `form:"sort"`
}

type DashboardTeacherHomeworkOverviewRequest struct {
	CourseID   int64 `form:"course_id"`
	HomeworkID int64 `form:"homework_id" binding:"required"`
	LessonID   int64 `form:"lesson_id"`
} 