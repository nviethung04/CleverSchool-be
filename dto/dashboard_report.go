package dto

type DashboardReportDTO struct {
	TotalSchools                 int64 `json:"total_schools"`
	TotalClasses                 int64 `json:"total_classes"`
	TotalUsers                   int64 `json:"total_users"`
	StudentActive                int64 `json:"student_active"`
	TeacherActive                int64 `json:"teacher_active"`
	StudentActiveWeekly          int64 `json:"student_active_weekly"`
	TeacherActiveWeekly          int64 `json:"teacher_active_weekly"`
	FrequentLoginStudents        int64 `json:"frequent_login_students"`
	FrequentLoginStudentsWeekly  int64 `json:"frequent_login_students_weekly"`
}

type DashboardReportResponseDTO struct {
	Data DashboardReportDTO `json:"data"`
}
