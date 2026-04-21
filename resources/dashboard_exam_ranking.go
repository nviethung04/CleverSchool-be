package resources

import (
	"be-lms/dto"
	"be-lms/prot"
	"be-lms/utils"
	"be-lms/models"
)

func DashboardExamRankingResource(ranking dto.DashboardExamRanking) *prot.DashboardExamRanking {
	return &prot.DashboardExamRanking{
		StudentId:    ranking.StudentID,
		StudentName:  ranking.StudentName,
		AverageRatio: ranking.AverageRatio,
		AverageTime:  ranking.AverageTime,
		Avatar:       utils.StaticURL(ranking.AvatarInfo.Path, models.Storage),
		NumberExams:  ranking.NumberExams,
	}
}

func DashboardExamRankingCollection(rankings []dto.DashboardExamRanking) []*prot.DashboardExamRanking {
	var result []*prot.DashboardExamRanking
	for _, ranking := range rankings {
		result = append(result, DashboardExamRankingResource(ranking))
	}
	return result
} 