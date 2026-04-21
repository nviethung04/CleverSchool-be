package dto

type DashboardStudentExamListItemDTO struct {
    ID           int64   `json:"id"`
    Name         string  `json:"name"`
    Description  string  `json:"description"`
    CoverImage   string  `json:"cover_image"`
    Deadline     int64   `json:"deadline"`
    IsSubmitted  bool    `json:"is_submitted"`
    UnscoredCount int    `json:"unscored_count"`
    Ratio        float64 `json:"ratio"`
    Duration     int64   `json:"duration"`
    CreatedAt    int64   `json:"created_at"`
    Week         int     `json:"week"`
    Year         int     `json:"year"`
    WeekStartDate int64  `json:"week_start_date"`
    WeekEndDate   int64  `json:"week_end_date"`
    CourseName   string  `json:"course_name"`
    LessonTitle  string  `json:"lesson_title"`
}

type DashboardStudentExamListResponseDTO struct {
    Exams []DashboardStudentExamListItemDTO `json:"exams"`
    Total int64                             `json:"total"`
} 