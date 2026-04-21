package dto

import "time"

// DashboardSchoolsResponse DTO cho response dashboard schools
type DashboardSchoolsResponse struct {
	ID                              int64     `json:"id"`
	SchoolID                        int64     `json:"school_id"`
	SchoolName                      string    `json:"school_name"`
	TotalStudents                   int64     `json:"total_students"`
	TotalTeachers                   int64     `json:"total_teachers"`
	ActiveTeachersFromSep8          int64     `json:"active_teachers_from_sep8"`
	ActiveStudentsFromSep15         int64     `json:"active_students_from_sep15"`
	StudentsCompletedHomeworkFromSep15 int64     `json:"students_completed_homework_from_sep15"`
	StudentsCompletedHomeworkSelectedWeek int64     `json:"students_completed_homework_selected_week"`
	ActiveStudentsSelectedWeek          int64     `json:"active_students_selected_week"`
	ActiveTeachersSelectedWeek          int64     `json:"active_teachers_selected_week"`
	StartDate                       time.Time `json:"start_date"`
	EndDate                         time.Time `json:"end_date"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}
