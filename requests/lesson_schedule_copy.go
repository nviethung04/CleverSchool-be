package requests

type CopyLessonScheduleRequest struct {
	SourceCourseID int64 `json:"source_course_id" form:"source_course_id" binding:"required"`
	TargetCourseID int64 `json:"target_course_id" form:"target_course_id" binding:"required"`
}
