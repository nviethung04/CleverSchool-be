package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
)

type ExamCommentService interface {
	CreateExamComment(examID, studentID, teacherID int64, content string) *prot.ExamCommentResponse
}

type examCommentService struct {
	repo repositories.ExamCommentRepository
}

func NewExamCommentService(repo repositories.ExamCommentRepository) ExamCommentService {
	return &examCommentService{repo: repo}
}

func (s *examCommentService) CreateExamComment(examID, studentID, teacherID int64, content string) *prot.ExamCommentResponse {
	comment := &models.ExamComment{
		ExamID:   examID,
		StudentID: studentID,
		TeacherID: teacherID,
		Content:   content,
	}
	err := s.repo.CreateExamComment(comment)
	if err != nil {
		return &prot.ExamCommentResponse{Success: false, Message: err.Error()}
	}
	return &prot.ExamCommentResponse{Success: true, Message: "Comment saved successfully"}
}
