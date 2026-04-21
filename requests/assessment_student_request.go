package requests

type AssessmentStudentRequest struct {
	CourseID     int64 `form:"course_id" binding:"required"`
	AssessmentID int64 `form:"assessment_id" binding:"required"`
	StudentID    int64 `form:"student_id"` // Filter theo student_id (optional)
	GetScore     bool  `form:"get_score"` // Nếu true, trả thêm điểm và chi tiết điểm
	Limit        int   `form:"limit"`
	Page         int   `form:"page"`
}


