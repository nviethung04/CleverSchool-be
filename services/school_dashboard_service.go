package services


import (
	"be-cleverschool/prot"
	"be-cleverschool/repositories"

	
)
type SchoolDashboardService interface {
GetAllSchoolsWithStats() ([]*prot.SchoolSummary, error)

}

type schoolDashboardService struct {
	repo repositories.SchoolDashboardRepository
}

func NewSchoolDashboardService(repo repositories.SchoolDashboardRepository) SchoolDashboardService {
	return &schoolDashboardService{repo: repo}
}

func (s *schoolDashboardService) GetAllSchoolsWithStats() ([]*prot.SchoolSummary, error) {
	data, err := s.repo.GetAllWithStats()
	if err != nil {
		return nil, err
	}

	var result []*prot.SchoolSummary
	for _, d := range data {
		result = append(result, &prot.SchoolSummary{
			Id:             d.ID,
			Name:           d.Name,
			StudentCount:   d.StudentCount,
			TeacherCount:   d.TeacherCount,
			ClassroomCount: d.ClassroomCount,
			CourseCount:    d.CourseCount,
		})
	}
	return result, nil
}

