package dto

type DashboardTeacherHomeworkListItem struct {
	HomeworkID         int64  `json:"homework_id"`
	HomeworkName       string `json:"homework_name"`
	TotalQuestions     int64  `json:"total_questions"`
	QuestionsCompleted int64  `json:"questions_completed"`
	Status             string `json:"status"` // "completed", "in_progress", "not_started"
	SubmittedAt        int64  `json:"submitted_at"` // Unix timestamp từ homework_users.updated_at
	AssignedAt         int64  `json:"assigned_at"`  // Unix timestamp từ homework_ref_lessons.assigned_at
	LessonID           int64  `json:"lesson_id"`
	LessonName         string `json:"lesson_name"`
	AssignLate         bool   `json:"assign_late"`
}

