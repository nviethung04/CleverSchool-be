package requests

type ContestStudentRequest struct {
	ContestRoundID int64 `form:"contest_round_id" json:"contest_round_id"`
	ContestID      int64 `form:"contest_id" json:"contest_id"`
	Limit          int   `form:"limit" json:"limit"`
	Page           int   `form:"page" json:"page"`
}

type GetContestByStudentRequest struct {
	UserID    int64 `form:"user_id" json:"user_id"`
	ContestID int64 `form:"contest_id" json:"contest_id"`
	Limit     int   `form:"limit" json:"limit"`
	Page      int   `form:"page" json:"page"`
}

type DashboardContestRequest struct {
	ContestID int64  `form:"contest_id" json:"contest_id"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
}

type DashboardContestListRequest struct {
	UserID    int64  `form:"user_id" json:"user_id"`
	ContestID int64  `form:"contest_id" json:"contest_id"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	Limit     int    `form:"limit" json:"limit"`
	Page      int    `form:"page" json:"page"`
}

type DashboardContestRoundRequest struct {
	ContestRoundID int64  `form:"contest_round_id" json:"contest_round_id"`
	StartDate      string `form:"start_date" json:"start_date"`
	EndDate        string `form:"end_date" json:"end_date"`
}

type DashboardContestRoundListRequest struct {
	UserID         int64  `form:"user_id" json:"user_id"`
	ContestID      int64  `form:"contest_id" json:"contest_id"`
	ContestRoundID int64  `form:"contest_round_id" json:"contest_round_id"`
	StartDate      string `form:"start_date" json:"start_date"`
	EndDate        string `form:"end_date" json:"end_date"`
	Limit          int    `form:"limit" json:"limit"`
	Page           int    `form:"page" json:"page"`
}
