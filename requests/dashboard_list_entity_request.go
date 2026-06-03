package requests

type DashboardSchoolListRequest struct {
	Page  int               `json:"page" form:"page"`
	Limit int               `json:"limit" form:"limit"`
	Sort  map[string]string `json:"sort" form:"sort"`
}

type DashboardCourseListRequest struct {
	Page     int               `json:"page" form:"page"`
	Limit    int               `json:"limit" form:"limit"`
	Sort     map[string]string `json:"sort" form:"sort"`
	SchoolID int64             `json:"school_id" form:"school_id"`
	ProgramID int64             `json:"program_id" form:"program_id"`
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
	LessonID    int64             `json:"lesson_id" form:"lesson_id"`
	IsAssigned  *bool             `json:"is_assigned" form:"is_assigned"`
}

type DashboardLessonListRequest struct {
	Page      int               `json:"page" form:"page"`
	Limit     int               `json:"limit" form:"limit"`
	Sort      map[string]string `json:"sort" form:"sort"`
	CourseID  int64             `json:"course_id" form:"course_id"`
	ChapterID int64             `json:"chapter_id" form:"chapter_id"`
}
