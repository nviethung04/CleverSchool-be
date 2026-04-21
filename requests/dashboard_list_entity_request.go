package requests

type DashboardSchoolListRequest struct {
	Page  int               `json:"page" form:"page"`
	Limit int               `json:"limit" form:"limit"`
	Sort  map[string]string `json:"sort" form:"sort"`
}

type DashboardCourseListRequest struct {
	Page           int               `json:"page" form:"page"`
	Limit          int               `json:"limit" form:"limit"`
	Sort           map[string]string `json:"sort" form:"sort"`
	SchoolID       int64             `json:"school_id" form:"school_id"`
	ProgramID      int64             `json:"program_id" form:"program_id"`
	SubjectID      int64             `json:"subject_id" form:"subject_id"`
	ParentCourseID *int64            `json:"parent_course_id" form:"parent_course_id"`
	AssessmentID   *int64            `json:"assessment_id" form:"assessment_id"`
}

type DashboardTeacherListRequest struct {
	Page     int               `json:"page" form:"page"`
	Limit    int               `json:"limit" form:"limit"`
	Sort     map[string]string `json:"sort" form:"sort"`
	CourseID int64             `json:"course_id" form:"course_id"`
	SchoolID int64             `json:"school_id" form:"school_id"`
	ProgramID int64             `json:"program_id" form:"program_id"`
}

type DashboardSubjectListRequest struct {
	Page  int               `json:"page" form:"page"`
	Limit int               `json:"limit" form:"limit"`
	Sort  map[string]string `json:"sort" form:"sort"`
}

type DashboardExamListRequest struct {
	Page      int               `json:"page" form:"page"`
	Limit     int               `json:"limit" form:"limit"`
	Sort      map[string]string `json:"sort" form:"sort"`
	CourseID  int64             `json:"course_id" form:"course_id"`
	SubjectID int64             `json:"subject_id" form:"subject_id"`
	LessonID  int64             `json:"lesson_id" form:"lesson_id"`
}

type DashboardHomeworkListRequest struct {
	Page        int               `json:"page" form:"page"`
	Limit       int               `json:"limit" form:"limit"`
	Sort        map[string]string `json:"sort" form:"sort"`
	CourseID    int64             `json:"course_id" form:"course_id"`
	SubjectID   int64             `json:"subject_id" form:"subject_id"`
	ChapterID   int64             `json:"chapter_id" form:"chapter_id"`
	LessonID    int64             `json:"lesson_id" form:"lesson_id"`
	LessonIDs   string            `json:"lesson_ids" form:"lesson_ids"`   // Danh sách lesson_id cách nhau bởi dấu phẩy (ví dụ: "1,2,3")
	HomeworkIDs string            `json:"homework_ids" form:"homework_ids"` // Danh sách homework_id cách nhau bởi dấu phẩy (ví dụ: "1,2,3")
	IsAssigned  *bool             `json:"is_assigned" form:"is_assigned"`
	StartDate   string            `json:"start_date" form:"start_date"` // Format: YYYY-MM-DD hoặc Unix timestamp
	EndDate     string            `json:"end_date" form:"end_date"`     // Format: YYYY-MM-DD hoặc Unix timestamp
}

type DashboardLessonListRequest struct {
	Page      int               `json:"page" form:"page"`
	Limit     int               `json:"limit" form:"limit"`
	Sort      map[string]string `json:"sort" form:"sort"`
	CourseID  int64             `json:"course_id" form:"course_id"`
	ChapterID int64             `json:"chapter_id" form:"chapter_id"`
}

type DashboardChapterListRequest struct {
	Page     int               `json:"page" form:"page"`
	Limit    int               `json:"limit" form:"limit"`
	Sort     map[string]string `json:"sort" form:"sort"`
	CourseID int64             `json:"course_id" form:"course_id"`
}

type DashboardClassListRequest struct {
	Page         int               `json:"page" form:"page"`
	Limit        int               `json:"limit" form:"limit"`
	Sort         map[string]string `json:"sort" form:"sort"`
	SchoolID     int64             `json:"school_id" form:"school_id"`
	ClassMainID  int64             `json:"class_main_id" form:"class_main_id"`
}

type DashboardClassMainListRequest struct {
	Page     int               `json:"page" form:"page"`
	Limit    int               `json:"limit" form:"limit"`
	Sort     map[string]string `json:"sort" form:"sort"`
	SchoolID int64             `json:"school_id" form:"school_id"`
}