package requests

type CourseFamilyRequest struct {
	CourseID int64 `form:"course_id" json:"course_id" binding:"required"`
}

