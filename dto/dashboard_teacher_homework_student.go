package dto

import "be-Clever School/models"

type DashboardTeacherHomeworkStudent struct {
	StudentID         int64      `json:"student_id"`
	StudentName       string     `json:"student_name"`
	StudentAvatar     models.MediaInfo `json:"student_avatar"`
	TotalQuestions    int64      `json:"total_questions"`
	QuestionsCompleted int64     `json:"questions_completed"`
}

type DashboardTeacherHomeworkStudentStats struct {
	StudentID            int64      `json:"student_id"`
	StudentName          string     `json:"student_name"`
	StudentAvatar        models.MediaInfo `json:"student_avatar"`
	TotalHomework        int64      `json:"total_homework"`
	TotalAssignedHomeworks int64    `json:"total_assigned_homeworks"`
	InProgressHomework   int64      `json:"in_progress_homework"`
	CompletedHomework    int64      `json:"completed_homework"`
	NotStartedHomework   int64      `json:"not_started_homework"`
	AverageRatio         float64    `json:"average_ratio"`
	CompletedHomeworkRatio float64  `json:"completed_homework_ratio"` // completed_homework / total_assigned_homeworks
	TotalQuestions        int64     `json:"total_questions"`
	QuestionsCompleted    int64     `json:"questions_completed"`
}

type DashboardTeacherHomeworkOverview struct {
	NotStartedPercentage float64 `json:"not_started_percentage"`
	InProgressPercentage float64 `json:"in_progress_percentage"`
	CompletedPercentage  float64 `json:"completed_percentage"`
	TotalStudents        int64   `json:"total_students"`
	NotStartedCount      int64   `json:"not_started_count"`
	InProgressCount      int64   `json:"in_progress_count"`
	CompletedCount       int64   `json:"completed_count"`
} 