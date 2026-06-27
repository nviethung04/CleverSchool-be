package dto

type ExamStudentItem struct {
	UserID        int64   `json:"user_id"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	Avatar        string  `json:"avatar"`
	IsSubmitted   bool    `json:"is_submitted"`
	SubmittedAt   int64   `json:"submitted_at"`
	Duration      int64   `json:"duration"`
	UnscoredCount int32   `json:"unscored_count"`
	Score         float64 `json:"score"`
	Ratio         float64 `json:"ratio"`
	CourseID      int64   `json:"course_id"`
}

type ExamStudentInfo struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	IsAssigned        bool    `json:"is_assigned"`
	Status            int32   `json:"status"`
	MaxScore          float64 `json:"max_score"`
	TimeLimit         int64   `json:"time_limit"`
	CreatedAt         int64   `json:"created_at"`
	Deadline          int64   `json:"deadline"`
	TotalQuestions    int32   `json:"total_questions"`
	TotalStudents     int32   `json:"total_students"`
}

type GetExamByStudentExamDTO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Status       int32   `json:"status"`
	Description  string  `json:"description"`
	CoverImage   string  `json:"cover_image"`
	Deadline     int64   `json:"deadline"`
	IsSubmitted  bool    `json:"is_submitted"`
	UnscoredCount int32  `json:"unscored_count"`
}

type GetExamByStudentHomeworkDTO struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	Status                 int32  `json:"status"`
	Description            string `json:"description"`
	TotalQuestion          int32  `json:"total_question"`
	QuestionCompleted      int32  `json:"question_completed"`
	CoverImage             string `json:"cover_image"`
	LastQuestionIDCompleted int64 `json:"last_question_id_completed"`
}

type GetExamByStudentExerciseDTO struct {
    ID            int64  `json:"id"`
    Name          string `json:"name"`
    Status        int32  `json:"status"`
    Description   string `json:"description"`
    CoverImage    string `json:"cover_image"`
    Deadline      int64  `json:"deadline"`
}

type GetExamByStudentAssessmentDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	IsSubmitted bool   `json:"is_submitted"`
	IsScored    bool   `json:"is_scored"`
	IsPublished bool   `json:"is_published"`
}

type GetExamByStudentLessonDTO struct {
	ID          int64                         `json:"id"`
	Title       string                        `json:"title"`
	Description string                        `json:"description"`
	Status      bool                          `json:"status"`
	Exams       []GetExamByStudentExamDTO     `json:"exams"`
	Homeworks   []GetExamByStudentHomeworkDTO `json:"homeworks"`
    Exercises   []GetExamByStudentExerciseDTO `json:"exercises"`
	Assessments []GetExamByStudentAssessmentDTO `json:"assessments"`
}

type GetExamByStudentCourseDTO struct {
	ID          int64                        `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Status      bool                         `json:"status"`
	SubjectName string                       `json:"subject_name"`
	Type        string                       `json:"type"`
	Image       string                       `json:"image"`
	Level       string                       `json:"level"`
	Target      string                       `json:"target"`
	Lessons     []GetExamByStudentLessonDTO  `json:"lessons"`
}

type GetExamByStudentResponseDTO struct {
	Courses []GetExamByStudentCourseDTO `json:"courses"`
	Total   int64                       `json:"total"`
}

type GetExamByStudentExamQuery struct {
	ID          int64
	Name        string
	Status      int32
	Description string
	CoverImage  string
	Deadline    int64
} 