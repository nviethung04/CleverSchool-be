package requests

type GetAssessmentRequest struct {
	Limit       int    `form:"limit"`
	Page        int    `form:"page"`
	Keyword     string `form:"keyword"`
	Type        string `form:"type"`
	ProgramID   *int64 `form:"program_id"`
	SubjectID   *int64 `form:"subject_id"`
	CourseID    *int64 `form:"course_id"`
	LessonID    *int64 `form:"lesson_id"`
	ClassMainID *int64 `form:"class_main_id"`
	GradeID     *int64 `form:"grade_id"`
}
