package dto

import "be-lms/models"

type DashboardExamRanking struct {
	StudentID    int64            `gorm:"column:student_id" json:"student_id"`
	StudentName  string           `gorm:"column:student_name" json:"student_name"`
	AverageRatio float64          `gorm:"column:average_ratio" json:"average_ratio"`
	AverageTime  float64          `gorm:"column:average_time" json:"average_time"`
	AvatarInfo   models.MediaInfo `gorm:"column:avatar_info" json:"avatar_info"`
	NumberExams  int64            `gorm:"column:number_exams" json:"number_exams"`
} 