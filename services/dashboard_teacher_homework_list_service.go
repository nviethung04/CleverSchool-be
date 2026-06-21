package services

import (
	"be-lms/repositories"
	"be-lms/requests"
	"strconv"
)

type DashboardTeacherHomeworkListService interface {
	GetStudentHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) (map[string]interface{}, error)
}

type dashboardTeacherHomeworkListService struct {
	repo repositories.DashboardTeacherHomeworkListRepository
}

func NewDashboardTeacherHomeworkListService() DashboardTeacherHomeworkListService {
	return &dashboardTeacherHomeworkListService{
		repo: repositories.NewDashboardTeacherHomeworkListRepository(),
	}
}

func homeworkListStatus(totalQuestions, questionsCompleted int64, hasSubmission bool) string {
	if !hasSubmission || questionsCompleted == 0 {
		return "not_started"
	}
	if totalQuestions > 0 && questionsCompleted >= totalQuestions {
		return "completed"
	}
	return "in_progress"
}

func (s *dashboardTeacherHomeworkListService) GetStudentHomeworkList(req *requests.DashboardTeacherHomeworkListRequest) (map[string]interface{}, error) {
	rows, total, err := s.repo.GetStudentHomeworkList(req)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]interface{}{
			"homework_id":         strconv.FormatInt(row.HomeworkID, 10),
			"homework_name":       row.HomeworkName,
			"total_questions":     strconv.FormatInt(row.TotalQuestions, 10),
			"questions_completed": strconv.FormatInt(row.QuestionsCompleted, 10),
			"status":              homeworkListStatus(row.TotalQuestions, row.QuestionsCompleted, row.HasSubmission),
			"submitted_at":        strconv.FormatInt(row.SubmittedAt, 10),
			"lesson_id":           strconv.FormatInt(row.LessonID, 10),
			"lesson_name":         row.LessonName,
			"assigned_at":         strconv.FormatInt(row.AssignedAt, 10),
			"assign_late":         row.AssignLate,
		})
	}

	return map[string]interface{}{
		"homeworks": items,
		"total":     total,
	}, nil
}
