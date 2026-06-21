package services

import (
	"be-lms/prot"
	"be-lms/repositories"
)

type DashboardStudentExerciseListService interface {
	GetStudentExerciseList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentHomeworkListResponse, error)
}

type dashboardStudentExerciseListService struct {
	repo repositories.DashboardStudentExerciseListRepository
}

func NewDashboardStudentExerciseListService(repo repositories.DashboardStudentExerciseListRepository) DashboardStudentExerciseListService {
	return &dashboardStudentExerciseListService{repo: repo}
}

func (s *dashboardStudentExerciseListService) GetStudentExerciseList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentHomeworkListResponse, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}
	items, total, err := s.repo.GetStudentExerciseList(userID, courseID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	var protoItems []*prot.DashboardStudentHomeworkListItem
	for _, h := range items {
		protoItems = append(protoItems, &prot.DashboardStudentHomeworkListItem{
			Id:                 h.ID,
			Name:               h.Name,
			Description:        h.Description,
			CoverImage:         h.CoverImage,
			CreatedAt:          h.CreatedAt,
			IsSubmitted:        h.IsSubmitted,
			AssignedAt:         h.AssignedAt,
			TotalQuestions:     int32(h.TotalQuestions),
			QuestionsCompleted: int32(h.QuestionsCompleted),
			Week:               int32(h.Week),
			Year:               int32(h.Year),
			WeekStartDate:      h.WeekStartDate,
			WeekEndDate:        h.WeekEndDate,
			CourseId:           h.CourseID,
			CourseName:         h.CourseName,
			LessonId:           h.LessonID,
			LessonTitle:        h.LessonTitle,
			Ratio:              h.Ratio,
			SkipQuestionsCount: int32(h.SkipQuestionsCount),
		})
	}
	return &prot.DashboardStudentHomeworkListResponse{
		Homeworks: protoItems,
		Total:     total,
	}, nil
}
