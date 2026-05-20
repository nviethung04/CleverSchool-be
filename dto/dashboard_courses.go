package dto

import (
	"be-cleverschool/models"
	"time"
)

// DashboardCoursesResponse DTO cho response dashboard courses
type DashboardCoursesResponse struct {
	ID                              int64     `json:"id"`
	CourseID                        int64     `json:"course_id"`
	CourseName                      string    `json:"course_name"`
	ObjectTitle                     string    `json:"object_title"`                     // Thêm object_title từ courses
	SchoolID                        int64     `json:"school_id"`
	SchoolName                      string    `json:"school_name"`
	TotalStudents                   int64     `json:"total_students"`
	TotalTeachers                   int64     `json:"total_teachers"`
	TeacherIDs                      string    `json:"teacher_ids"`
	TeacherInfos                    models.DashboardTeacherInfos `json:"teacher_infos"`
	ActiveTeachersFromSep8          int64     `json:"active_teachers_from_sep8"`
	ActiveStudentsFromSep15         int64     `json:"active_students_from_sep15"`
	StudentsCompletedHomeworkFromSep15 int64 `json:"students_completed_homework_from_sep15"`
	StudentsCompletedHomeworkSelectedWeek int64  `json:"students_completed_homework_selected_week"`
	ActiveStudentsSelectedWeek          int64     `json:"active_students_selected_week"`
	ActiveTeachersSelectedWeek          int64     `json:"active_teachers_selected_week"`
	TotalHomeworks                  int64     `json:"total_homeworks"`                   // Tổng số homework trong khoảng thời gian selected
	AssignedHomeworks               int64     `json:"assigned_homeworks"`               // Số homework đã giao trong khoảng thời gian selected
	StudentsCompletedAllHomeworks   int64     `json:"students_completed_all_homeworks"` // Số học sinh đã hoàn thành tất cả homework được giao trong khoảng thời gian selected
	StudentsDoingHomeworks          int64     `json:"students_doing_homeworks"`         // Số học sinh đang làm homework trong khoảng thời gian selected
	StudentsNotStartedAnyHomework    int64     `json:"students_not_started_any_homework"` // Số học sinh chưa làm homework nào trong khoảng thời gian selected
	StudentActiveNotStartedAnyHomework int64   `json:"student_active_not_started_any_homework"` // Số học sinh active nhưng chưa làm homework nào
	HomeworkOver50PercentStudentComplete int64 `json:"homework_over_50_percent_student_complete"` // Số lượng homework có trên 50% học sinh active hoàn thành
	StartDate                       time.Time `json:"start_date"`
	EndDate                         time.Time `json:"end_date"`
	CreatedAt                       time.Time `json:"created_at"`
	UpdatedAt                       time.Time `json:"updated_at"`
}

