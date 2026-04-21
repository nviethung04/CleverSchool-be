package controllers

import (
	"be-lms/prot"
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type HomeworkRankingController struct {
	svc services.HomeworkRankingService
}

func NewHomeworkRankingController() *HomeworkRankingController {
	return &HomeworkRankingController{
		svc: services.NewHomeworkRankingService(),
	}
}

func (ctl *HomeworkRankingController) GetHomeworkRanking(c *gin.Context) {
	var req requests.HomeworkRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	rankings, total, err := ctl.svc.GetHomeworkRanking(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	students := make([]*prot.RankingStudentItem, len(rankings))
	for i := range rankings {
		students[i] = &prot.RankingStudentItem{
			Id:    rankings[i].StudentID,
			Name:  rankings[i].StudentName,
			Avatar: rankings[i].Avatar,
			Class: rankings[i].Class,
			Star:  int32(rankings[i].Star),
			Rank:  int32(rankings[i].Rank),
			IsMe:  rankings[i].IsMe,
			Exp:   rankings[i].Exp,
		}
	}
	response := &prot.RankingListResponse{
		Students: students,
		Total:   total,
	}

	utils.Respond(c, response, nil, "")
}
