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

type HomeworkRankingService interface {
	GetHomeworkRanking(c *gin.Context, req *requests.HomeworkRankingRequest) ([]dto.HomeworkRanking, int64, error)
}

type homeworkRankingService struct {
	repo repositories.HomeworkRankingRepository
}

func NewHomeworkRankingService() HomeworkRankingService {
	return &homeworkRankingService{
		repo: repositories.NewHomeworkRankingRepository(),
	}
}

func (s *homeworkRankingService) GetHomeworkRanking(c *gin.Context, req *requests.HomeworkRankingRequest) ([]dto.HomeworkRanking, int64, error) {
	orderBy := req.OrderBy
	if orderBy != "star" && orderBy != "exp" {
		return nil, 0, errors.New("order_by phải là 'star' hoặc 'exp'")
	}

	rows, total, err := s.repo.GetHomeworkRanking(req)
	if err != nil {
		return nil, 0, err
	}

	currentUserID := int64(utils.GetCurrentUserId(c))

	result := make([]dto.HomeworkRanking, len(rows))
	for i := range rows {
		// Tính rank: tìm row đầu tiên có cùng giá trị (star và exp đều bằng nhau)
		rank := i + 1
		for j := i - 1; j >= 0; j-- {
			isTied := false
			if orderBy == "star" {
				// Nếu star và exp đều bằng nhau thì đồng hạng
				isTied = rows[i].Star == rows[j].Star && rows[i].Exp == rows[j].Exp
			} else {
				// Nếu exp và star đều bằng nhau thì đồng hạng
				isTied = rows[i].Exp == rows[j].Exp && rows[i].Star == rows[j].Star
			}
			if isTied {
				// Tìm thấy row đồng hạng, lấy rank của row đó
				rank = result[j].Rank
				break
			}
		}

		result[i] = dto.HomeworkRanking{
			StudentID:   rows[i].StudentID,
			StudentName: rows[i].StudentName,
			Avatar:      utils.StaticURL(rows[i].AvatarInfo.Path, models.Storage),
			Class:       rows[i].Class,
			Star:        rows[i].Star,
			Exp:         rows[i].Exp,
			Rank:        rank,
			IsMe:        currentUserID == rows[i].StudentID,
		}
	}
	return result, total, nil
}

