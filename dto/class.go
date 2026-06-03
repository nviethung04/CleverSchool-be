package dto

import "be-lms/models"

type Student struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	IsActivity   bool   `json:"is_activity"`
	ActivityDate string `json:"activity_date"`
}

type Course struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Class struct {
	TotalStudent   int32 `json:"total_student"`
	ActiveStudents int32 `json:"active_students"`
	Class          models.Class
	Courses        []Course  `json:"courses"`
	Students       []Student `json:"students"`
}
