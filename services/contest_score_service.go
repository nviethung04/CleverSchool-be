package services

import (
	"be-cleverschool/repositories"

	"github.com/gin-gonic/gin"
)

type ContestScoreService interface {
	SaveContestScoreMultipleChoice(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreFillInBlank(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreOrderingDragdrop(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreMatching(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreLabeling(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreCategory(c *gin.Context, req interface{}) (interface{}, error)
	SaveContestScoreManualScoring(c *gin.Context, req interface{}) (interface{}, error)
	SubmitContestRound(c *gin.Context, req interface{}) (interface{}, error)
	SkipContestQuestion(c *gin.Context, req interface{}) (interface{}, error)
	CheckSubmitContestRound(c *gin.Context, contestRoundId int64) (interface{}, error)
}

type contestScoreService struct {
	repo repositories.ContestScoreRepository
}

func NewContestScoreService() ContestScoreService {
	return &contestScoreService{
		repo: repositories.NewContestScoreRepository(),
	}
}

func (s *contestScoreService) SaveContestScoreMultipleChoice(c *gin.Context, req interface{}) (interface{}, error) {
	// Implementation similar to exam/homework multiple choice scoring
	// but using contest_round_question_user_multiple_choices table
	return s.repo.SaveContestScoreMultipleChoice(req)
}

func (s *contestScoreService) SaveContestScoreFillInBlank(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreFillInBlank(req)
}

func (s *contestScoreService) SaveContestScoreOrderingDragdrop(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreOrderingDragdrop(req)
}

func (s *contestScoreService) SaveContestScoreMatching(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreMatching(req)
}

func (s *contestScoreService) SaveContestScoreLabeling(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreLabeling(req)
}

func (s *contestScoreService) SaveContestScoreCategory(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreCategory(req)
}

func (s *contestScoreService) SaveContestScoreManualScoring(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SaveContestScoreManualScoring(req)
}

func (s *contestScoreService) SubmitContestRound(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SubmitContestRound(req)
}

func (s *contestScoreService) SkipContestQuestion(c *gin.Context, req interface{}) (interface{}, error) {
	return s.repo.SkipContestQuestion(req)
}

func (s *contestScoreService) CheckSubmitContestRound(c *gin.Context, contestRoundId int64) (interface{}, error) {
	return s.repo.CheckSubmitContestRound(contestRoundId)
}

