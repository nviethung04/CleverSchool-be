package requests

type AssessmentExportExcelRequest struct {
	CourseIDs    string `form:"course_ids" binding:"required"` // Dạng "1,2,3,4"
	AssessmentID int64  `form:"assessment_id" binding:"required"`
}

