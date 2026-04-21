package requests

type ExerciseStudentRequest struct {
    ExerciseID int64 `form:"exercise_id"`
    CourseID   int64 `form:"course_id"`
    Limit      int   `form:"limit"`
    Page       int   `form:"page"`
}


