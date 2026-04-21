package dto

import "time"

type DashboardTeacherHomeworkUnscored struct {
	UserID             int64     `json:"user_id"`
	HomeworkID         int64     `json:"homework_id"`
	CourseID           int64     `json:"course_id"`
	LessonID           int64     `json:"lesson_id"`
	StudentName        string    `json:"student_name"`
	CourseName         string    `json:"course_name"`
	ObjectTitle        string    `json:"object_title"`
	SubjectName        string    `json:"subject_name"`
	HomeworkName       string    `json:"homework_name"`
	LessonTitle        string    `json:"lesson_title"`
	SubmittedAt        time.Time `json:"submitted_at"`
	Ratio              float64   `json:"ratio"`
	TotalQuestions     int       `json:"total_questions"`
	UnscoredQuestions  int       `json:"unscored_questions"`
	ManualQuestions    int       `json:"manual_questions"`
	HasComment         bool      `json:"has_comment"`
	CommentContent     string    `json:"comment_content"`
}
