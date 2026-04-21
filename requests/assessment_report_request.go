package requests

// AssessmentReportRequest dùng cho API /api/dashboard/assessment/report
// Query params:
// - class_id (required nếu class_main_id không có)
// - class_main_id (optional, nếu > 0 thì bỏ qua class_id và lấy học sinh từ tất cả classes có class_main_id này)
// - assessment_id (required)
// - student_id (optional, filter theo 1 hoặc nhiều học sinh, có thể truyền nhiều lần: ?student_id=1&student_id=2)
// - get_score (optional, nếu true thì trả thêm điểm & chi tiết điểm)
// - get_study_report_criteria (optional, nếu true thì trả thêm study report theo assessments.study_report_criteria_id)
// - limit, page (optional, phân trang)
type AssessmentReportRequest struct {
	ClassID                 int64   `form:"class_id"`                      // Required nếu class_main_id không có
	ClassMainID            int64   `form:"class_main_id"`                 // Nếu > 0 thì bỏ qua class_id
	AssessmentID           int64   `form:"assessment_id" binding:"required"`
	StudentID              []int64 `form:"student_id"`                    // Filter theo student_id (optional, có thể truyền nhiều)
	GetScore               bool    `form:"get_score"`                     // Nếu true, trả thêm điểm và chi tiết điểm
	GetStudyReportCriteria bool    `form:"get_study_report_criteria"`     // Nếu true, trả thêm study report
	Limit                  int     `form:"limit"`
	Page                   int     `form:"page"`
}


