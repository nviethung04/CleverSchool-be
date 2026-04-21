package dto

import "time"

type HomeworkStudentDoing struct {
	HomeworkID  int64
	UserID      int64
	Username    string
	Name        string
	SchoolName  string
	SubmittedAt time.Time
}

