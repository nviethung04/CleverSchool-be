package resources

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
)

func DashboardTeacherExamOverviewResource(overview dto.DashboardTeacherExamOverview) *prot.DashboardTeacherExamOverview {
	return &prot.DashboardTeacherExamOverview{
		TotalSubmitted: overview.TotalSubmitted,
		TotalScored:    overview.TotalScored,
		TotalUnscored:  overview.TotalUnscored,
		CompletionRate: overview.CompletionRate,
	}
} 
