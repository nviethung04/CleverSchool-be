package repositories

import (
	"gorm.io/gorm"
)

type ContestQuestionRepository interface {
	GetContestRoundQuestions(contestRoundId int64) (interface{}, error)
	AssignQuestionsToContestRound(contestRoundId int64, questionIds []int64) error
	RemoveQuestionsFromContestRound(contestRoundId int64, questionIds []int64) error
}

type contestQuestionRepository struct{}

func NewContestQuestionRepository() ContestQuestionRepository {
	return &contestQuestionRepository{}
}

func (r *contestQuestionRepository) GetContestRoundQuestions(contestRoundId int64) (interface{}, error) {
	// This method is now handled by cloned_questions table
	// The actual implementation is in contest_question_service.go
	return nil, nil
}

func (r *contestQuestionRepository) AssignQuestionsToContestRound(contestRoundId int64, questionIds []int64) error {
	// This method is now handled by cloned_questions table
	// The actual implementation is in contest_question_service.go
	return nil
}

func (r *contestQuestionRepository) RemoveQuestionsFromContestRound(contestRoundId int64, questionIds []int64) error {
	// This method is now handled by cloned_questions table
	// The actual implementation is in contest_question_service.go
	return nil
}

// ContestRoundQuestion model (if not exists)
type ContestRoundQuestion struct {
	ID             int64 `gorm:"primaryKey"`
	ContestRoundId int64
	QuestionId     int64
	CreatedAt      *gorm.DeletedAt
	CreatedBy      int64
	UpdatedAt      *gorm.DeletedAt
	UpdatedBy      int64
	DeletedAt      *gorm.DeletedAt
	DeletedBy      int64
}
