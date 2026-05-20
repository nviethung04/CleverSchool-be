package services

import (
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/utils"
	"errors"

	"github.com/gin-gonic/gin"
)

type AssessmentRankingService interface {
	GetAssessmentRanking(c *gin.Context, req *requests.AssessmentRankingRequest) ([]dto.AssessmentRanking, int64, error)
}

type assessmentRankingService struct {
	repo repositories.AssessmentRankingRepository
}

func NewAssessmentRankingService() AssessmentRankingService {
	return &assessmentRankingService{
		repo: repositories.NewAssessmentRankingRepository(),
	}
}

func (s *assessmentRankingService) GetAssessmentRanking(c *gin.Context, req *requests.AssessmentRankingRequest) ([]dto.AssessmentRanking, int64, error) {
	orderBy := req.OrderBy
	if orderBy != "score" && orderBy != "star" {
		return nil, 0, errors.New("order_by phải là 'score' hoặc 'star'")
	}

	rows, total, err := s.repo.GetAssessmentRanking(req)
	if err != nil {
		return nil, 0, err
	}

	currentUserID := int64(utils.GetCurrentUserId(c))

	result := make([]dto.AssessmentRanking, len(rows))
	for i := range rows {
		// Tính rank: tìm row đầu tiên có cùng giá trị (star và score đều bằng nhau)
		rank := i + 1
		for j := i - 1; j >= 0; j-- {
			isTied := false
			if orderBy == "star" {
				// Nếu star và score đều bằng nhau thì đồng hạng
				isTied = rows[i].Star == rows[j].Star && rows[i].Score == rows[j].Score
			} else {
				// Nếu score và star đều bằng nhau thì đồng hạng
				isTied = rows[i].Score == rows[j].Score && rows[i].Star == rows[j].Star
			}
			if isTied {
				// Tìm thấy row đồng hạng, lấy rank của row đó
				rank = result[j].Rank
				break
			}
		}

		result[i] = dto.AssessmentRanking{
			StudentID:   rows[i].StudentID,
			StudentName: rows[i].StudentName,
			Avatar:      utils.StaticURL(rows[i].AvatarInfo.Path, models.Storage),
			Class:       rows[i].Class,
			Score:       rows[i].Score,
			Star:        rows[i].Star,
			Rank:        rank,
			IsMe:        currentUserID == rows[i].StudentID,
		}
	}
	return result, total, nil
}

