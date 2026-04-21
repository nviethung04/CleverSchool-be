package requests

type AssessmentRefLessonRequest struct {
	AssessmentID int64 `json:"assessment_id" binding:"required"`
	LessonID     int64 `json:"lesson_id" binding:"required"`
	CourseID     int64 `json:"course_id" binding:"required"`
}
