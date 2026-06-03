package dto

import "be-lms/models"

type DashboardContestRanking struct {
	StudentID    int64            `json:"student_id"`
	StudentName  string           `json:"student_name"`
	AverageRatio float64          `json:"average_ratio"`
	AverageScore float64          `json:"average_score"`
	AverageTime  float64          `json:"average_time"` // Thời gian trung bình (giây)
	AvatarInfo   models.MediaInfo `gorm:"column:avatar_info" json:"avatar_info"`
	NumberRounds int64            `gorm:"column:number_rounds" json:"number_rounds"`
}

type ContestScoreChart struct {
	ScoreRange   string `json:"score_range"`
	StudentCount int64  `json:"student_count"`
}

