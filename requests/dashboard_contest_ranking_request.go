package requests

type DashboardContestRankingRequest struct {
	CourseID       int64  `form:"course_id"`
	ContestID      int64  `form:"contest_id"`
	ContestRoundID int64  `form:"contest_round_id"`
	ClassID        int64  `form:"class_id"`        // Filter theo lớp
	SchoolID       int64  `form:"school_id"`       // Filter theo trường
	ProvinceCode   string `form:"province_code"`   // Filter theo tỉnh/thành phố
	StartDate      *int64 `form:"start_date"`
	EndDate        *int64 `form:"end_date"`
	Limit          int    `form:"limit"`
	Page           int    `form:"page"`
	OrderBy        string `form:"order_by"` // "asc" or "desc"
}
