package requests
 
type DashboardStudentExamRequest struct {
    CourseID  *int64 `form:"course_id"`
    StartDate *int64 `form:"start_date"`
    EndDate   *int64 `form:"end_date"`
} 