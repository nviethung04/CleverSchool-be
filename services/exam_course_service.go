package services

import (
	"be-cleverschool/repositories"
)

type ExamCourseService interface {
	GetExamCourseDetail(courseID int64, isAssigned *bool) (repositories.ExamCourseDetail, error)
}

type examCourseService struct {
	repo repositories.ExamCourseRepository
}

func NewExamCourseService(repo repositories.ExamCourseRepository) ExamCourseService {
	return &examCourseService{repo: repo}
}

func (s *examCourseService) GetExamCourseDetail(courseID int64, isAssigned *bool) (repositories.ExamCourseDetail, error) {
	return s.repo.GetExamCourseDetail(courseID, isAssigned)
} 
