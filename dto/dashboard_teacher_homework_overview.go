package dto

type DashboardTeacherHomeworkOverviewStats struct {
	TotalSubmitted         int64   `json:"total_submitted"`
	TotalScored            int64   `json:"total_scored"`
	TotalUnscored          int64   `json:"total_unscored"`
	TotalNoManualScoring   int64   `json:"total_no_manual_scoring"`
	CompletionRate         float64 `json:"completion_rate"`
}