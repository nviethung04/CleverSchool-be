package services

import (
	"be-lms/repositories"
	"be-lms/requests"
	"strconv"
)

type DashboardTeacherExerciseListService interface {
	GetStudentExerciseList(req *requests.DashboardTeacherExerciseListRequest) (map[string]interface{}, error)
}

type dashboardTeacherExerciseListService struct {
	repo repositories.DashboardTeacherExerciseListRepository
}

func NewDashboardTeacherExerciseListService() DashboardTeacherExerciseListService {
	return &dashboardTeacherExerciseListService{
		repo: repositories.NewDashboardTeacherExerciseListRepository(),
	}
}

func exerciseListStatus(totalQuestions, questionsCompleted int64, hasSubmission bool) string {
	if !hasSubmission || questionsCompleted == 0 {
		return "not_started"
	}
	if totalQuestions > 0 && questionsCompleted >= totalQuestions {
		return "completed"
	}
	return "in_progress"
}

func (s *dashboardTeacherExerciseListService) GetStudentExerciseList(req *requests.DashboardTeacherExerciseListRequest) (map[string]interface{}, error) {
	rows, total, err := s.repo.GetStudentExerciseList(req)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]interface{}{
			"exercise_id":         strconv.FormatInt(row.ExerciseID, 10),
			"exercise_name":       row.ExerciseName,
			"total_questions":     strconv.FormatInt(row.TotalQuestions, 10),
			"questions_completed": strconv.FormatInt(row.QuestionsCompleted, 10),
			"status":              exerciseListStatus(row.TotalQuestions, row.QuestionsCompleted, row.HasSubmission),
			"ratio":               row.Ratio,
			"submitted_at":        strconv.FormatInt(row.SubmittedAt, 10),
			"lesson_id":           strconv.FormatInt(row.LessonID, 10),
			"lesson_name":         row.LessonName,
			"assigned_at":         strconv.FormatInt(row.AssignedAt, 10),
			"assign_late":         row.AssignLate,
		})
	}

	return map[string]interface{}{
		"exercises": items,
		"total":     total,
	}, nil
}
