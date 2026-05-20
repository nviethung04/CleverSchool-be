package resources

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
)

func DashboardContestRankingCollection(rankings []dto.DashboardContestRanking) []*prot.DashboardContestRanking {
	var result []*prot.DashboardContestRanking
	for _, ranking := range rankings {
		result = append(result, &prot.DashboardContestRanking{
			StudentId:    ranking.StudentID,
			StudentName:  ranking.StudentName,
			AverageRatio: ranking.AverageRatio,
			AverageScore: ranking.AverageScore,
			AverageTime:  ranking.AverageTime,
			Avatar:       ranking.AvatarInfo.Path,
			NumberRounds: ranking.NumberRounds,
		})
	}
	return result
}

