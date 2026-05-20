package services

import (
	"be-cleverschool/repositories"

	"github.com/gin-gonic/gin"
)

type ContestResultService interface {
	GetContestRoundAnswers(c *gin.Context, contestRoundId int64, userId int64) (interface{}, error)
	GetContestRoundAnswersByStudent(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundResults(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundLeaderboard(c *gin.Context, contestRoundId int64) (interface{}, error)
}

type contestResultService struct {
	repo repositories.ContestResultRepository
}

func NewContestResultService() ContestResultService {
	return &contestResultService{
		repo: repositories.NewContestResultRepository(),
	}
}

func (s *contestResultService) GetContestRoundAnswers(c *gin.Context, contestRoundId int64, userId int64) (interface{}, error) {
	answers, err := s.repo.GetContestRoundAnswers(contestRoundId, userId)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *contestResultService) GetContestRoundAnswersByStudent(c *gin.Context, contestRoundId int64) (interface{}, error) {
	answers, err := s.repo.GetContestRoundAnswersByStudent(contestRoundId)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (s *contestResultService) GetContestRoundResults(c *gin.Context, contestRoundId int64) (interface{}, error) {
	results, err := s.repo.GetContestRoundResults(contestRoundId)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *contestResultService) GetContestRoundLeaderboard(c *gin.Context, contestRoundId int64) (interface{}, error) {
	leaderboard, err := s.repo.GetContestRoundLeaderboard(contestRoundId)
	if err != nil {
		return nil, err
	}
	return leaderboard, nil
}

