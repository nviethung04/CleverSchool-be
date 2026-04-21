package dto

import "time"

type DashboardTeacherExamUnscored struct {
	UserID             int64     `json:"user_id"`
	ExamID             int64     `json:"exam_id"`
	CourseID           int64     `json:"course_id"`
	LessonID           int64     `json:"lesson_id"`
	StudentName        string    `json:"student_name"`
	CourseName         string    `json:"course_name"`
	SubjectName        string    `json:"subject_name"`
	ExamName           string    `json:"exam_name"`
	LessonTitle        string    `json:"lesson_title"`
	SubmittedAt        time.Time `json:"submitted_at"`
	Deadline           time.Time `json:"deadline"`
	SubmitOnTime       bool      `json:"submit_on_time"`
	Ratio              float64   `json:"ratio"`
	TotalQuestions     int       `json:"total_questions"`
	UnscoredQuestions  int       `json:"unscored_questions"`
} 