package requests

type CourseScheduleAssignParentRequest struct {
	CourseIDs      string `json:"course_ids" form:"course_ids"`
	ParentCourseID int64  `json:"parent_course_id" form:"parent_course_id" binding:"required"`
}

type CourseScheduleSyncFamilyRequest struct {
	CourseID int64 `json:"course_id" form:"course_id" binding:"required"`
}

