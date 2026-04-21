package dto

type CopyLessonScheduleRequest struct {
	SourceCourseID int64 `json:"source_course_id" binding:"required"`
	TargetCourseID int64 `json:"target_course_id" binding:"required"`
}

type CopyLessonScheduleResponse struct {
	Success              bool   `json:"success"`
	Message              string `json:"message"`
	CopiedSchedulesCount int32  `json:"copied_schedules_count"`
}
