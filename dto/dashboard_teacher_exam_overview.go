package dto

type DashboardTeacherExamOverview struct {
	TotalSubmitted    int64   `json:"total_submitted"`
	TotalScored       int64   `json:"total_scored"`
	TotalUnscored     int64   `json:"total_unscored"`
	CompletionRate    float64 `json:"completion_rate"`
} 