package services

import (
	"be-lms/repositories"
)

type HomeworkStudentService interface {
	GetHomeworkStudents(homeworkID int64, courseID int64) (repositories.HomeworkInfo, []repositories.HomeworkStudentInfo, error)
}

type homeworkStudentService struct {
	repo repositories.HomeworkStudentRepository
}

func NewHomeworkStudentService(repo repositories.HomeworkStudentRepository) HomeworkStudentService {
	return &homeworkStudentService{repo: repo}
}

func (s *homeworkStudentService) GetHomeworkStudents(homeworkID int64, courseID int64) (repositories.HomeworkInfo, []repositories.HomeworkStudentInfo, error) {
	return s.repo.GetHomeworkStudents(homeworkID, courseID)
} 