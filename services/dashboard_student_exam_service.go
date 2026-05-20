package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
)

type DashboardStudentExamService interface {
	GetStudentExamStats(userID int64, courseID *int64, startDate, endDate *int64) (*prot.DashboardStudentExamResponse, error)
}

type dashboardStudentExamService struct {
	repo repositories.DashboardStudentExamRepository
}

func NewDashboardStudentExamService(repo repositories.DashboardStudentExamRepository) DashboardStudentExamService {
	return &dashboardStudentExamService{repo: repo}
}

func (s *dashboardStudentExamService) GetStudentExamStats(userID int64, courseID *int64, startDate, endDate *int64) (*prot.DashboardStudentExamResponse, error) {
	repoResult, err := s.repo.GetStudentExamStats(userID, courseID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var chart []*prot.DashboardStudentExamWeek
	for _, w := range repoResult.Chart {
		chart = append(chart, &prot.DashboardStudentExamWeek{
			WeekNumber:    int32(w.WeekNumber),
			Year:          int32(w.Year),
			ExamSubmitted: int32(w.ExamSubmitted),
			AvgScore:      w.AvgScore,
			StartDate:     w.StartDate,
			EndDate:       w.EndDate,
			AvgRatio:      w.AvgRatio,
			ExamAssigned:  int32(w.ExamAssigned),
		})
	}
	overview := &prot.DashboardStudentExamOverview{
		TotalExamAssigned:  int32(repoResult.Overview.TotalExamAssigned),
		TotalExamSubmitted: int32(repoResult.Overview.TotalExamSubmitted),
		AvgRatio:           repoResult.Overview.AvgRatio,
	}
	return &prot.DashboardStudentExamResponse{
		Overview: overview,
		Chart:    chart,
	}, nil
}

