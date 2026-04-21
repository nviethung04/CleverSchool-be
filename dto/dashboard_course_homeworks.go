package dto

import "time"

type DashboardCourseHomework struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	AssignedAt     *time.Time `json:"assigned_at"`
	IsAssignedLate bool       `json:"is_assigned_late"`
}

type DashboardCourseHomeworksResponse struct {
	Homeworks []DashboardCourseHomework `json:"homeworks"`
}

