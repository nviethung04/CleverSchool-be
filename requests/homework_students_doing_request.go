package requests

type HomeworkStudentsDoingRequest struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Keyword    string `form:"keyword"`
	HomeworkID int64  `form:"homework_id" binding:"required"`
}

