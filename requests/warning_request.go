package requests

type GetActiveUsersRequest struct {
	Minutes   int    `form:"minutes"`     // Số phút để check hoạt động
	StartTime string `form:"start_time"`  // Unix timestamp bắt đầu (optional, dùng khi minutes=0)
	EndTime   string `form:"end_time"`    // Unix timestamp kết thúc (optional, dùng khi minutes=0)
	Limit     int    `form:"limit"`       // Số lượng records mỗi page
	Page      int    `form:"page"`        // Trang hiện tại
	SchoolID  int64  `form:"school_id"`    // Filter theo school_id (optional)
	CourseID  int64  `form:"course_id"`   // Filter theo course_id (optional)
	ProgramID int64  `form:"program_id"`  // Filter theo program_id (optional)
}

type GetFailedLoginRequest struct {
	Minutes   int    `form:"minutes"`    // Số phút để check login lỗi (default 60)
	StartTime string `form:"start_time"` // Unix timestamp bắt đầu (optional, dùng khi minutes=0)
	EndTime   string `form:"end_time"`   // Unix timestamp kết thúc (optional, dùng khi minutes=0)
	UserID    uint   `form:"user_id"`    // Filter theo user cụ thể (optional)
	Limit     int    `form:"limit"`      // Số lượng records mỗi page
	Page      int    `form:"page"`       // Trang hiện tại
}
