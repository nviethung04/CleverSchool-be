package services

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"math"
)

type DashboardSchoolsProtoService interface {
	GetDashboardSchoolsProto(req *prot.DashboardSchoolsRequest) (*prot.DashboardSchoolsResponse, error)
}

type dashboardSchoolsProtoService struct {
	dashboardSchoolsRepo repositories.DashboardSchoolsRepository
}

func NewDashboardSchoolsProtoService(dashboardSchoolsRepo repositories.DashboardSchoolsRepository) DashboardSchoolsProtoService {
	return &dashboardSchoolsProtoService{
		dashboardSchoolsRepo: dashboardSchoolsRepo,
	}
}

// GetDashboardSchoolsProto lấy danh sách dashboard schools với protobuf
func (s *dashboardSchoolsProtoService) GetDashboardSchoolsProto(req *prot.DashboardSchoolsRequest) (*prot.DashboardSchoolsResponse, error) {
	var schools []dto.DashboardSchoolsResponse
	var total int64
	var err error

	// Nếu không truyền page và limit thì lấy hết
	if req.Page <= 0 && req.Limit <= 0 {
		// Lấy tất cả schools
		schools, err = s.dashboardSchoolsRepo.GetDashboardSchools(req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}
		total = int64(len(schools))
	} else {
		// Set default values nếu có truyền page hoặc limit
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}

		// Get schools with pagination and search
		schools, err = s.dashboardSchoolsRepo.GetDashboardSchoolsWithPagination(req.Search, int(req.Page), int(req.Limit), req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}

		// Get total count for pagination
		total, err = s.dashboardSchoolsRepo.GetDashboardSchoolsCount(req.Search)
		if err != nil {
			return nil, err
		}
	}

	// Convert to protobuf format
	var protoSchools []*prot.DashboardReportSchool
	for _, school := range schools {
		protoSchool := &prot.DashboardReportSchool{
			Id:                                    school.ID,
			SchoolId:                              school.SchoolID,
			SchoolName:                            school.SchoolName,
			TotalStudents:                         school.TotalStudents,
			TotalTeachers:                         school.TotalTeachers,
			ActiveTeachersFromSep8:                school.ActiveTeachersFromSep8,
			ActiveStudentsFromSep15:               school.ActiveStudentsFromSep15,
			StudentsCompletedHomeworkFromSep15:    school.StudentsCompletedHomeworkFromSep15,
			StudentsCompletedHomeworkSelectedWeek:     school.StudentsCompletedHomeworkSelectedWeek,
			ActiveStudentsSelectedWeek:                school.ActiveStudentsSelectedWeek,
			ActiveTeachersSelectedWeek:                school.ActiveTeachersSelectedWeek,
		}
		protoSchools = append(protoSchools, protoSchool)
	}

	// Calculate total pages
	var totalPages int32
	if req.Limit > 0 {
		totalPages = int32(math.Ceil(float64(total) / float64(req.Limit)))
	} else {
		totalPages = 1 // Nếu không có limit thì chỉ có 1 trang
	}

	return &prot.DashboardSchoolsResponse{
		Schools:     protoSchools,
		Total:       total,
		Page:        req.Page,
		Limit:       req.Limit,
		TotalPages:  totalPages,
	}, nil
}

