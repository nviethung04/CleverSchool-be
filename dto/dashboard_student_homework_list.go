package dto

type DashboardStudentHomeworkListItemDTO struct {
    ID                int64   `json:"id"`
    Name              string  `json:"name"`
    Description       string  `json:"description"`
    CoverImage        string  `json:"cover_image"`
    CreatedAt         int64   `json:"created_at"`
    IsSubmitted       bool    `json:"is_submitted"`
    TotalQuestions    int     `json:"total_questions"`
    QuestionsCompleted int    `json:"questions_completed"`
    Week              int     `json:"week"`
    Year              int     `json:"year"`
    WeekStartDate     int64   `json:"week_start_date"`
    WeekEndDate       int64   `json:"week_end_date"`
    AssignedAt       int64   `json:"assigned_at"`
    CourseID          int64   `json:"course_id"`
    CourseName        string  `json:"course_name"`
    LessonID          int64   `json:"lesson_id"`
    LessonTitle       string  `json:"lesson_title"`
    Ratio             float64 `json:"ratio"`
    SkipQuestionsCount int    `json:"skip_questions_count"`
}

type DashboardStudentHomeworkListResponseDTO struct {
    Homeworks []DashboardStudentHomeworkListItemDTO `json:"homeworks"`
    Total     int64                                `json:"total"`
}
