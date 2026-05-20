package services

import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
)

type DashboardStudentExamListService interface {
	GetStudentExamList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentExamListResponse, error)
}

type dashboardStudentExamListService struct {
	repo repositories.DashboardStudentExamListRepository
}

func NewDashboardStudentExamListService(repo repositories.DashboardStudentExamListRepository) DashboardStudentExamListService {
	return &dashboardStudentExamListService{repo: repo}
}

func (s *dashboardStudentExamListService) GetStudentExamList(userID int64, courseID *int64, startDate, endDate *int64, limit, page int) (*prot.DashboardStudentExamListResponse, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}
	exams, total, err := s.repo.GetStudentExamList(userID, courseID, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	var protoExams []*prot.DashboardStudentExamListItem
	for _, e := range exams {
		protoExams = append(protoExams, &prot.DashboardStudentExamListItem{
			Id:           e.ID,
			Name:         e.Name,
			Description:  e.Description,
			CoverImage:   e.CoverImage,
			Deadline:     e.Deadline,
			IsSubmitted:  e.IsSubmitted,
			UnscoredCount: int32(e.UnscoredCount),
			Ratio:        e.Ratio,
			Duration:     e.Duration,
			CreatedAt:    e.CreatedAt,
			Week:         int32(e.Week),
			Year:         int32(e.Year),
			WeekStartDate: e.WeekStartDate,
			WeekEndDate:   e.WeekEndDate,
			CourseName:    e.CourseName,
			LessonTitle:   e.LessonTitle,
		})
	}
	return &prot.DashboardStudentExamListResponse{
		Exams: protoExams,
		Total: total,
	}, nil
}

