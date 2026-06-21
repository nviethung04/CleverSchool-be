package services

import (
	"be-lms/repositories"
	"be-lms/requests"
	"strconv"
)

type DashboardTeacherExerciseStudentService interface {
	GetStudentStats(req *requests.DashboardTeacherExerciseStudentStatsRequest) (map[string]interface{}, error)
}

type dashboardTeacherExerciseStudentService struct {
	repo repositories.DashboardTeacherExerciseStudentRepository
}

func NewDashboardTeacherExerciseStudentService() DashboardTeacherExerciseStudentService {
	return &dashboardTeacherExerciseStudentService{
		repo: repositories.NewDashboardTeacherExerciseStudentRepository(),
	}
}

func (s *dashboardTeacherExerciseStudentService) GetStudentStats(req *requests.DashboardTeacherExerciseStudentStatsRequest) (map[string]interface{}, error) {
	rows, total, err := s.repo.GetStudentStats(req)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]interface{}{
			"student_id":               strconv.FormatInt(row.StudentID, 10),
			"student_name":             row.StudentName,
			"total_exercises":          strconv.FormatInt(row.TotalExercises, 10),
			"total_assigned_exercises": strconv.FormatInt(row.TotalAssignedExercises, 10),
			"in_progress_exercises":    strconv.FormatInt(row.InProgressExercises, 10),
			"completed_exercises":      strconv.FormatInt(row.CompletedExercises, 10),
			"not_started_exercises":    strconv.FormatInt(row.NotStartedExercises, 10),
			"average_ratio":            row.AverageRatio,
		})
	}

	return map[string]interface{}{
		"students": items,
		"total":    total,
	}, nil
}
