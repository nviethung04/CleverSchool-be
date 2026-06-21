package dto

type DashboardTeacherHomeworkListItem struct {
	HomeworkID         int64  `gorm:"column:homework_id"`
	HomeworkName       string `gorm:"column:homework_name"`
	TotalQuestions     int64  `gorm:"column:total_questions"`
	QuestionsCompleted int64  `gorm:"column:questions_completed"`
	HasSubmission      bool   `gorm:"column:has_submission"`
	SubmittedAt        int64  `gorm:"column:submitted_at"`
	LessonID           int64  `gorm:"column:lesson_id"`
	LessonName         string `gorm:"column:lesson_name"`
	AssignedAt         int64  `gorm:"column:assigned_at"`
	AssignLate         bool   `gorm:"column:assign_late"`
}
