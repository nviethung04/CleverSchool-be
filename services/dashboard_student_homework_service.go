package services

import (
	"be-Clever School/prot"
	"be-Clever School/repositories"
)

type DashboardStudentHomeworkService interface {
	GetStudentHomeworkStats(userID int64, courseID *int64, startDate, endDate *int64) (*prot.DashboardStudentHomeworkResponse, error)
}

type dashboardStudentHomeworkService struct {
	repo repositories.DashboardStudentHomeworkRepository
}

func NewDashboardStudentHomeworkService(repo repositories.DashboardStudentHomeworkRepository) DashboardStudentHomeworkService {
	return &dashboardStudentHomeworkService{repo: repo}
}

func (s *dashboardStudentHomeworkService) GetStudentHomeworkStats(userID int64, courseID *int64, startDate, endDate *int64) (*prot.DashboardStudentHomeworkResponse, error) {
	repoResult, err := s.repo.GetStudentHomeworkStats(userID, courseID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	var chart []*prot.DashboardStudentHomeworkWeek
	for _, w := range repoResult.Chart {
		chart = append(chart, &prot.DashboardStudentHomeworkWeek{
			WeekNumber:         int32(w.WeekNumber),
			Year:               int32(w.Year),
			CourseWeek:         int32(w.CourseWeek),
			HomeworkSubmitted:  int32(w.HomeworkSubmitted),
			StartDate:          w.StartDate,
			EndDate:            w.EndDate,
			HomeworkAssigned:   int32(w.HomeworkAssigned),
			QuestionsCompleted: int32(w.QuestionsCompleted),
			TotalQuestions:     int32(w.TotalQuestions),
			PercentCompleted:   w.PercentCompleted,
			AverageRatio:       w.AverageRatio,
		})
	}
	overview := &prot.DashboardStudentHomeworkOverview{
		TotalHomeworkAssigned:   int32(repoResult.Overview.TotalHomeworkAssigned),
		TotalHomeworkSubmitted:  int32(repoResult.Overview.TotalHomeworkSubmitted),
		TotalQuestionsCompleted: int32(repoResult.Overview.TotalQuestionsCompleted),
		TotalQuestions:          int32(repoResult.Overview.TotalQuestions),
		PercentCompleted:        repoResult.Overview.PercentCompleted,
		AverageRatio:            repoResult.Overview.AverageRatio,
	}
	return &prot.DashboardStudentHomeworkResponse{
		Overview: overview,
		Chart:    chart,
	}, nil
}
