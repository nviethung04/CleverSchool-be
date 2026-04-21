package dto

type DashboardStudentExamWeekDTO struct {
    WeekNumber    int     `json:"week_number"`
    Year          int     `json:"year"`
    ExamSubmitted int     `json:"exam_submitted"`
    AvgScore      float64 `json:"avg_score"`
    StartDate     int64   `json:"start_date"`
    EndDate       int64   `json:"end_date"`
    AvgRatio      float64 `json:"avg_ratio"`
    ExamAssigned  int     `json:"exam_assigned"`
}

type DashboardStudentExamOverviewDTO struct {
    TotalExamAssigned  int     `json:"total_exam_assigned"`
    TotalExamSubmitted int     `json:"total_exam_submitted"`
    AvgRatio          float64 `json:"avg_ratio"`
}

type DashboardStudentExamResponseDTO struct {
    Overview DashboardStudentExamOverviewDTO         `json:"overview"`
    Chart    []DashboardStudentExamWeekDTO           `json:"chart"`
} 