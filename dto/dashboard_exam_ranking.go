package dto

import "be-lms/models"

type DashboardExamRanking struct {
	StudentID    int64            `json:"student_id"`
	StudentName  string           `json:"student_name"`
	AverageRatio float64          `json:"average_ratio"`
	AverageTime  float64          `json:"average_time"`  // Thời gian trung bình (giây)
	AvatarInfo   models.MediaInfo `gorm:"column:avatar_info" json:"avatar_info"`
	NumberExams  int64            `gorm:"column:number_exams" json:"number_exams"`
} 