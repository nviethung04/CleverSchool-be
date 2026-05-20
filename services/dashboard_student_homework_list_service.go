package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
)

type DashboardStudentHomeworkListService interface {
    GetStudentHomeworkList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentHomeworkListResponse, error)
}

type dashboardStudentHomeworkListService struct {
    repo repositories.DashboardStudentHomeworkListRepository
}

func NewDashboardStudentHomeworkListService(repo repositories.DashboardStudentHomeworkListRepository) DashboardStudentHomeworkListService {
    return &dashboardStudentHomeworkListService{repo: repo}
}

func (s *dashboardStudentHomeworkListService) GetStudentHomeworkList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentHomeworkListResponse, error) {
    offset := 0
    if page > 0 && limit > 0 {
        offset = (page - 1) * limit
    }
    homeworks, total, err := s.repo.GetStudentHomeworkList(userID, courseID, startDate, endDate, limit, offset)
    if err != nil {
        return nil, err
    }
    var protoHomeworks []*prot.DashboardStudentHomeworkListItem
    for _, h := range homeworks {
        protoHomeworks = append(protoHomeworks, &prot.DashboardStudentHomeworkListItem{
            Id:                h.ID,
            Name:              h.Name,
            Description:       h.Description,
            CoverImage:        h.CoverImage,
            CreatedAt:         h.CreatedAt,
            IsSubmitted:       h.IsSubmitted,
            AssignedAt: h.AssignedAt,
            TotalQuestions:    int32(h.TotalQuestions),
            QuestionsCompleted: int32(h.QuestionsCompleted),
            Week:              int32(h.Week),
            Year:              int32(h.Year),
            WeekStartDate:     h.WeekStartDate,
            WeekEndDate:       h.WeekEndDate,
            CourseId:          h.CourseID,
            CourseName:        h.CourseName,
            LessonId:          h.LessonID,
            LessonTitle:       h.LessonTitle,
            Ratio:             h.Ratio,
            SkipQuestionsCount: int32(h.SkipQuestionsCount),
            QuestionForm:      h.QuestionForm,
            Rate:              h.Rate,
        })
    }
    return &prot.DashboardStudentHomeworkListResponse{
        Homeworks: protoHomeworks,
        Total:     total,
    }, nil
}

