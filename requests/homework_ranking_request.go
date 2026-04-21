package requests

type HomeworkRankingRequest struct {
	CourseID int64  `form:"course_id" binding:"required"`
	OrderBy  string `form:"order_by" binding:"required"` // "star" hoặc "exp"
	Sort     string `form:"sort"`                        // "asc" hoặc "desc" (mặc định là "desc")
}
