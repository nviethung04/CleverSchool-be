package resources

import (
	"be-lms/dto"
	"be-lms/prot"
)

func DashboardTeacherExamOverviewResource(overview dto.DashboardTeacherExamOverview) *prot.DashboardTeacherExamOverview {
	return &prot.DashboardTeacherExamOverview{
		TotalSubmitted: overview.TotalSubmitted,
		TotalScored:    overview.TotalScored,
		TotalUnscored:  overview.TotalUnscored,
		CompletionRate: overview.CompletionRate,
	}
} 