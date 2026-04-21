package dto

// DashboardCourseStudentHomework DTO cho homework của học sinh
type DashboardCourseStudentHomework struct {
	ID              int64 `json:"id"`
	Name            string `json:"name"`
	IsCompleted     bool   `json:"is_completed"` // Trạng thái hoàn thành
	IsCompletedLate bool   `json:"is_completed_late"` // Bài làm sau end_date
}

// DashboardCourseStudent DTO cho học sinh trong course
type DashboardCourseStudent struct {
	ID        int64                             `json:"id"`
	Name      string                            `json:"name"`
	IsActive  bool                              `json:"is_active"` // Trạng thái active (true/false)
	Homeworks []DashboardCourseStudentHomework `json:"homeworks"` // Danh sách homework
}

// DashboardCourseStudentsResponse DTO cho response
type DashboardCourseStudentsResponse struct {
	Students []DashboardCourseStudent `json:"students"`
}

