package resources

import (
	"be-lms/dto"
	"be-lms/prot"
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