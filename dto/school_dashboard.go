// File: dto/school_dto.go
package dto

type SchoolSummaryResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	StudentCount   int64  `json:"student_count"`
	TeacherCount   int64  `json:"teacher_count"`
	ClassroomCount int64  `json:"classroom_count"`
	CourseCount    int64  `json:"course_count"` 
}