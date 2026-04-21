package requests

// GetFeedbackRequest - Request parameters cho API get all feedbacks
type GetFeedbackRequest struct {
	UserID      *int64 `form:"user_id"`      // Filter theo user_id
	RoleID      *int64 `form:"role_id"`      // Filter theo role_id
	Status      *int64 `form:"status"`       // Filter theo status
	Type        *int64 `form:"type"`         // Filter theo type
	StartDate   *int64 `form:"start_date"`   // Timestamp bắt đầu (chỉ so sánh ngày)
	EndDate     *int64 `form:"end_date"`     // Timestamp kết thúc (chỉ so sánh ngày)
	OrderBy     string `form:"order_by"`     // "newest" hoặc "oldest"
	Limit       int    `form:"limit"`        // Số lượng records per page
	Page        int    `form:"page"`         // Số trang
	Keyword     string `form:"keyword"`      // Tìm kiếm theo content
}
