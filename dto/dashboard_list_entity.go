package dto

import "time"

type DashboardSchool struct {
	ID              int64  `json:"id"`
	SchoolName      string `json:"school_name"`
	SchoolShortName string `json:"school_short_name"`
	WardName        string `json:"ward_name"`
	ProvinceName    string `json:"province_name"`
}

type DashboardCourse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	ObjectTitle    string `json:"object_title"`
	ParentCourseID int64  `json:"parent_course_id"`
	ProgramID      int64  `json:"program_id"`
	ProgramName    string `json:"program_name"`
}

type DashboardTeacher struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DashboardSubject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type DashboardExam struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CourseID  int64  `json:"course_id"`
	SubjectID int64  `json:"subject_id"`
}

type DashboardHomework struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	CourseID     int64      `json:"course_id"`
	SubjectID    int64      `json:"subject_id"`
	CourseName   string     `json:"course_name"`
	SubjectName  string     `json:"subject_name"`
	LessonID     int64      `json:"lesson_id"`
	LessonTitle  string     `json:"lesson_title"`
	AssignedAt   *time.Time `json:"assigned_at,omitempty"`
}

type DashboardLesson struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ChapterID   int64  `json:"chapter_id"`
	ChapterName string `json:"chapter_name"`
	CourseID    int64  `json:"course_id"`
	CourseName  string `json:"course_name"`
}

type DashboardChapter struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	CourseID   int64  `json:"course_id"`
	CourseName string `json:"course_name"`
}

type DashboardClass struct {
	ID            int64  `json:"id"`
	ClassName     string `json:"class_name"`
	ClassMainID   int64  `json:"class_main_id"`
	ClassMainName string `json:"class_main_name"`
}

type DashboardClassMain struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	SchoolID int64  `json:"school_id"`
} 