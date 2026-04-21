package requests

type GetExerciseRequest struct {
    LessonID   *int  `form:"lesson_id"`
    IsAssigned *bool `form:"is_assigned"`
    Limit      int   `form:"limit"`
    Page       int   `form:"page"`
}


