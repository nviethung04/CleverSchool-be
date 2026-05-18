package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func DashboardTeacherExamOverviewResource(overview dto.DashboardTeacherExamOverview) *prot.DashboardTeacherExamOverview {
	return &prot.DashboardTeacherExamOverview{
		TotalSubmitted: overview.TotalSubmitted,
		TotalScored:    overview.TotalScored,
		TotalUnscored:  overview.TotalUnscored,
		CompletionRate: overview.CompletionRate,
	}
} 