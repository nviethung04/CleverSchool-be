package requests

type AssessmentRankingRequest struct {
	AssessmentID int64  `form:"assessment_id" binding:"required"`
	CourseID     int64  `form:"course_id" binding:"required"`
	OrderBy      string `form:"order_by" binding:"required"` // "score" hoặc "star"
	Sort         string `form:"sort"`                        // "asc" hoặc "desc" (mặc định "desc")
}
