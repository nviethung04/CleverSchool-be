package services

import (
    "be-lms/models"
    "be-lms/prot"
    "be-lms/repositories"
)

type HomeworkCommentService interface {
    CreateHomeworkComment(homeworkID, studentID, teacherID int64, content string) *prot.ExamCommentResponse
}

type homeworkCommentService struct {
    repo repositories.HomeworkCommentRepository
}

func NewHomeworkCommentService(repo repositories.HomeworkCommentRepository) HomeworkCommentService {
    return &homeworkCommentService{repo: repo}
}

func (s *homeworkCommentService) CreateHomeworkComment(homeworkID, studentID, teacherID int64, content string) *prot.ExamCommentResponse {
    comment := &models.HomeworkComment{
        HomeworkID: homeworkID,
        StudentID:  studentID,
        TeacherID:  teacherID,
        Content:    content,
    }
    if err := s.repo.CreateHomeworkComment(comment); err != nil {
        return &prot.ExamCommentResponse{Success: false, Message: err.Error()}
    }
    return &prot.ExamCommentResponse{Success: true, Message: "Comment saved successfully"}
}


