package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func DashboardTeacherHomeworkOverviewStatsResource(overview dto.DashboardTeacherHomeworkOverviewStats) *prot.DashboardTeacherHomeworkOverviewStats {
	return &prot.DashboardTeacherHomeworkOverviewStats{
		TotalSubmitted:         overview.TotalSubmitted,
		TotalScored:            overview.TotalScored,
		TotalUnscored:          overview.TotalUnscored,
		TotalNoManualScoring:   overview.TotalNoManualScoring,
		CompletionRate:         overview.CompletionRate,
	}
}