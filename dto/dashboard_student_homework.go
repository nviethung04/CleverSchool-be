package dto

type DashboardStudentHomeworkWeekDTO struct {
    WeekNumber        int     `json:"week_number"`
    Year              int     `json:"year"`
    CourseWeek        int     `json:"course_week"`
    HomeworkSubmitted int     `json:"homework_submitted"`
    StartDate         int64   `json:"start_date"`
    EndDate           int64   `json:"end_date"`
    HomeworkAssigned  int     `json:"homework_assigned"`
    QuestionsCompleted int    `json:"questions_completed"`
    TotalQuestions    int     `json:"total_questions"`
    PercentCompleted  float64 `json:"percent_completed"`
    AverageRatio      float64 `json:"average_ratio"`
}

type DashboardStudentHomeworkOverviewDTO struct {
    TotalHomeworkAssigned    int     `json:"total_homework_assigned"`
    TotalHomeworkSubmitted   int     `json:"total_homework_submitted"`
    TotalQuestionsCompleted  int     `json:"total_questions_completed"`
    TotalQuestions           int     `json:"total_questions"`
    PercentCompleted         float64 `json:"percent_completed"`
    AverageRatio             float64 `json:"average_ratio"`
}

type DashboardStudentHomeworkResponseDTO struct {
    Overview DashboardStudentHomeworkOverviewDTO         `json:"overview"`
    Chart    []DashboardStudentHomeworkWeekDTO           `json:"chart"`
} 