package dto

import "time"

// ContestRoundResultDTO - DTO tối ưu cho kết quả contest round
type ContestRoundResultDTO struct {
	ID                 int64   `json:"id"`
	ContestRoundID     int64   `json:"contest_round_id"`
	UserID             int64   `json:"user_id"`
	Score              float64 `json:"score"`
	Ratio              float64 `json:"ratio"`
	Time               int64   `json:"time"`
	HasManualScoring   bool    `json:"has_manual_scoring"`
	CreatedAt          time.Time `json:"created_at"`
	
	// User info - chỉ các trường cần thiết
	UserName           string  `json:"user_name"`
	UserCode           string  `json:"user_code"`
	UserPhone          string  `json:"user_phone"`
	UserEmail          string  `json:"user_email"`
	
	// School info - chỉ tên trường
	SchoolName         string  `json:"school_name"`
	
	// Address info - địa chỉ tỉnh/thành phố
	ProvinceName       string  `json:"province_name"`
	
	// Contest Round info - chỉ tên
	ContestRoundName   string  `json:"contest_round_name"`
	ContestName        string  `json:"contest_name"`
}

// ContestRoundLeaderboardDTO - DTO cho bảng xếp hạng
type ContestRoundLeaderboardDTO struct {
	Rank               int     `json:"rank"`
	UserID             int64   `json:"user_id"`
	UserName           string  `json:"user_name"`
	UserCode           string  `json:"user_code"`
	SchoolName         string  `json:"school_name"`
	Score              float64 `json:"score"`
	Ratio              float64 `json:"ratio"`
	Time               int64   `json:"time"`
}