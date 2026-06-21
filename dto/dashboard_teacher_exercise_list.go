package dto

type DashboardTeacherExerciseListItem struct {
	ExerciseID         int64   `gorm:"column:exercise_id"`
	ExerciseName       string  `gorm:"column:exercise_name"`
	TotalQuestions     int64   `gorm:"column:total_questions"`
	QuestionsCompleted int64   `gorm:"column:questions_completed"`
	HasSubmission      bool    `gorm:"column:has_submission"`
	SubmittedAt        int64   `gorm:"column:submitted_at"`
	Ratio              float64 `gorm:"column:ratio"`
	LessonID           int64   `gorm:"column:lesson_id"`
	LessonName         string  `gorm:"column:lesson_name"`
	AssignedAt         int64   `gorm:"column:assigned_at"`
	AssignLate         bool    `gorm:"column:assign_late"`
}

type DashboardTeacherExerciseStudentStats struct {
	StudentID              int64   `json:"student_id"`
	StudentName            string  `json:"student_name"`
	TotalExercises         int64   `json:"total_exercises"`
	TotalAssignedExercises int64   `json:"total_assigned_exercises"`
	InProgressExercises    int64   `json:"in_progress_exercises"`
	CompletedExercises     int64   `json:"completed_exercises"`
	NotStartedExercises    int64   `json:"not_started_exercises"`
	AverageRatio           float64 `json:"average_ratio"`
}
