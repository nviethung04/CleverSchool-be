package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type ContestService interface {
	GetAll(c *gin.Context) ([]*models.Contest, int64, error)
	GetByID(c *gin.Context, id int) (*models.Contest, error)
	Create(c *gin.Context, req *prot.ContestRequest) (*models.Contest, error)
	Update(c *gin.Context, req *prot.ContestRequest) (*models.Contest, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Contest, error)
	GetContestRounds(c *gin.Context, contestId int64) ([]*models.ContestRound, error)
	GetContestWithRounds(c *gin.Context, contestId int64) (*models.Contest, error)
	GetUserContests(c *gin.Context) ([]*models.Contest, int64, error)
}

type contestService struct {
	repo repositories.ContestRepository
}

func NewContestService(repo repositories.ContestRepository) ContestService {
	return &contestService{repo: repo}
}

func (s *contestService) GetAll(c *gin.Context) ([]*models.Contest, int64, error) {
	s.repo.SetContext(c)
	contests, total, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	// Convert slice of structs to slice of pointers and preload Creator
	result := make([]*models.Contest, len(contests))
	for i := range contests {
		result[i] = &contests[i]
		// Preload Creator for each contest
		if result[i].CreatedBy > 0 {
			var creator models.User
			if err := db.ReplicaDB.Where("id = ?", result[i].CreatedBy).First(&creator).Error; err == nil {
				result[i].Creator = &creator
			}
		}
	}

	return result, total, err
}

func (s *contestService) GetByID(c *gin.Context, id int) (*models.Contest, error) {
	s.repo.SetContext(c)
	contest, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *contestService) Create(c *gin.Context, req *prot.ContestRequest) (*models.Contest, error) {
	contestResource := resources.NewContestResource()
	contest := contestResource.FormatModelContest(req)
	contest.CreatedBy = int64(utils.GetCurrentUserId(c))
	contest.CreatedAt = time.Now()
	contest.UpdatedAt = time.Now()

	s.repo.SetContext(c)
	err := s.repo.Create(contest)
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *contestService) Update(c *gin.Context, req *prot.ContestRequest) (*models.Contest, error) {
	contestResource := resources.NewContestResource()
	contest := contestResource.FormatModelContest(req)
	contest.UpdatedBy = int64(utils.GetCurrentUserId(c))
	contest.UpdatedAt = time.Now()

	s.repo.SetContext(c)
	err := s.repo.Update(contest)
	if err != nil {
		return nil, err
	}
	return contest, nil
}

func (s *contestService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *contestService) Restore(c *gin.Context, id int) (*models.Contest, error) {
	s.repo.SetContext(c)
	return s.repo.Restore(id)
}

func (s *contestService) GetContestRounds(c *gin.Context, contestId int64) ([]*models.ContestRound, error) {
	rounds, err := s.repo.GetContestRounds(contestId)
	if err != nil {
		return nil, err
	}

	// Convert slice of structs to slice of pointers
	result := make([]*models.ContestRound, len(rounds))
	for i := range rounds {
		result[i] = &rounds[i]
	}

	return result, nil
}

func (s *contestService) GetContestWithRounds(c *gin.Context, contestId int64) (*models.Contest, error) {
	contest, err := s.repo.GetContestWithRounds(contestId)
	if err != nil {
		return nil, err
	}
	return contest, nil
}

// Helper function to convert model to protobuf
func modelToProtoContest(contest *models.Contest) *prot.Contest {
	protoContest := &prot.Contest{
		Id:          contest.ID,
		Name:        contest.Name,
		Description: contest.Description,
		Status:      int32(contest.Status),
		CreatedAt:   contest.CreatedAt.Unix(),
		CreatedBy:   contest.CreatedBy,
		UpdatedAt:   contest.UpdatedAt.Unix(),
		UpdatedBy:   contest.UpdatedBy,
		DeletedAt:   contest.DeletedAt.Time.Unix(),
		DeletedBy:   contest.DeletedBy,
	}

	if contest.StartTime != nil {
		protoContest.StartTime = contest.StartTime.Unix()
	}
	if contest.EndTime != nil {
		protoContest.EndTime = contest.EndTime.Unix()
	}

	// Convert contest rounds
	for _, round := range contest.ContestRounds {
		protoRound := &prot.ContestRound{
			Id:           round.ID,
			ContestId:    round.ContestId,
			Name:         round.Name,
			Description:  round.Description,
			SortPosition: round.SortPosition,
			JoinLevel:    round.JoinLevel,
			CreatedAt:    round.CreatedAt.Unix(),
			CreatedBy:    round.CreatedBy,
			UpdatedAt:    round.UpdatedAt.Unix(),
			UpdatedBy:    round.UpdatedBy,
			DeletedAt:    round.DeletedAt.Time.Unix(),
			DeletedBy:    round.DeletedBy,
		}

		if round.StartTime != nil {
			protoRound.StartTime = round.StartTime.Unix()
		}
		if round.EndTime != nil {
			protoRound.EndTime = round.EndTime.Unix()
		}

		protoContest.ContestRounds = append(protoContest.ContestRounds, protoRound)
	}

	return protoContest
}

// GetUserContests implementation
func (s *contestService) GetUserContests(c *gin.Context) ([]*models.Contest, int64, error) {
	s.repo.SetContext(c)
	contests, total, err := s.repo.GetUserContests()
	if err != nil {
		return nil, 0, err
	}

	// Convert slice of structs to slice of pointers
	result := make([]*models.Contest, len(contests))
	for i := range contests {
		result[i] = &contests[i]
	}

	return result, total, err
}
