package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
)

type ExerciseCommentService interface {
    CreateExerciseComment(exerciseID, studentID, teacherID int64, content string) *prot.ExamCommentResponse
}

type exerciseCommentService struct {
    repo repositories.ExerciseCommentRepository
}

func NewExerciseCommentService(repo repositories.ExerciseCommentRepository) ExerciseCommentService {
    return &exerciseCommentService{repo: repo}
}

func (s *exerciseCommentService) CreateExerciseComment(exerciseID, studentID, teacherID int64, content string) *prot.ExamCommentResponse {
    comment := &models.ExerciseComment{
        ExerciseID: exerciseID,
        StudentID:   studentID,
        TeacherID:   teacherID,
        Content:     content,
    }
    if err := s.repo.CreateExerciseComment(comment); err != nil {
        return &prot.ExamCommentResponse{Success: false, Message: err.Error()}
    }
    return &prot.ExamCommentResponse{Success: true, Message: "Comment saved successfully"}
}


