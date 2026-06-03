package services

import (
	"be-lms/prot"
	"be-lms/repositories"
)

type ContestStudentService interface {
	GetContestRoundsByStudent(userID int64, limit, page int) (*prot.ContestRoundListResponse, error)
}

type contestStudentService struct {
	repo repositories.ContestStudentRepository
}

func NewContestStudentService(repo repositories.ContestStudentRepository) ContestStudentService {
	return &contestStudentService{repo: repo}
}

func (s *contestStudentService) GetContestRoundsByStudent(userID int64, limit, page int) (*prot.ContestRoundListResponse, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}

	contestRounds, total, err := s.repo.GetContestRoundsByStudent(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	var result []*prot.ContestRound
	for _, cr := range contestRounds {
		result = append(result, &prot.ContestRound{
			Id:          cr.ID,
			Name:        cr.Name,
			Description: cr.Description,
			StartTime:   cr.StartTime,
			EndTime:     cr.EndTime,
		})
	}

	return &prot.ContestRoundListResponse{
		ContestRounds: result,
		TotalCount:    total,
	}, nil
}
