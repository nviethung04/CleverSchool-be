package services

import (
	"be-lms/repositories"
)

type ClonedQuestionService interface {
	GetQuestionsMap(assignmentID int64, assignmentType string) (map[string]repositories.ClonedQuestion, error)
}

type clonedQuestionService struct {
	repo repositories.ClonedQuestionRepository
}

func NewClonedQuestionService(repo repositories.ClonedQuestionRepository) ClonedQuestionService {
	return &clonedQuestionService{repo: repo}
}

func (s *clonedQuestionService) GetQuestionsMap(assignmentID int64, assignmentType string) (map[string]repositories.ClonedQuestion, error) {
	questions, err := s.repo.GetClonedQuestions(assignmentID, assignmentType)
	if err != nil {
		return nil, err
	}
	result := make(map[string]repositories.ClonedQuestion)
	for _, q := range questions {
		result[q.ID.String()] = q
	}
	return result, nil
}
