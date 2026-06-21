package requests

type ResetExerciseAttemptRequest struct {
	ExerciseID int64 `json:"exercise_id" binding:"required"`
	UserID     int64 `json:"user_id" binding:"required"`
	LessonID   int64 `json:"lesson_id"`
}
