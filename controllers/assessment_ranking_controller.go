package controllers

import (
	"be-lms/prot"
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type AssessmentRankingController struct {
	svc services.AssessmentRankingService
}

func NewAssessmentRankingController() *AssessmentRankingController {
	return &AssessmentRankingController{
		svc: services.NewAssessmentRankingService(),
	}
}

// GetAssessmentRanking godoc
// @Summary Bảng xếp hạng assessment theo score/star
// @Description Trả về danh sách học sinh trong khóa với score (assessment_scores), star (study_reports), is_me theo token.
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param assessment_id query int true "Assessment ID"
// @Param course_id query int true "Course ID"
// @Param order_by query string true "score hoặc star"
// @Param sort query string false "asc hoặc desc (mặc định desc)"
// @Success 200 {object} map[string]interface{}
// @Router /api/dashboard/ranking/assessment [get]
// @Security ApiKeyAuth
func (ctl *AssessmentRankingController) GetAssessmentRanking(c *gin.Context) {
	var req requests.AssessmentRankingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	rankings, total, err := ctl.svc.GetAssessmentRanking(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	students := make([]*prot.RankingStudentItem, len(rankings))
	for i := range rankings {
		students[i] = &prot.RankingStudentItem{
			Id:     rankings[i].StudentID,
			Name:   rankings[i].StudentName,
			Avatar: rankings[i].Avatar,
			Class:  rankings[i].Class,
			Star:   int32(rankings[i].Star),
			Rank:   int32(rankings[i].Rank),
			IsMe:   rankings[i].IsMe,
			Score:  rankings[i].Score,
		}
	}
	response := &prot.RankingListResponse{
		Students: students,
		Total:    total,
	}

	utils.Respond(c, response, nil, "")
}
