package services

import (
	"be-lms/dto"
	"be-lms/prot"
	"be-lms/repositories"
	"math"
)

type DashboardCoursesProtoService interface {
	GetDashboardCoursesProto(req *prot.DashboardCoursesRequest) (*prot.DashboardCoursesResponse, error)
}

type dashboardCoursesProtoService struct {
	dashboardCoursesRepo repositories.DashboardCoursesRepository
}

func NewDashboardCoursesProtoService(dashboardCoursesRepo repositories.DashboardCoursesRepository) DashboardCoursesProtoService {
	return &dashboardCoursesProtoService{
		dashboardCoursesRepo: dashboardCoursesRepo,
	}
}

// GetDashboardCoursesProto lấy danh sách dashboard courses với protobuf
func (s *dashboardCoursesProtoService) GetDashboardCoursesProto(req *prot.DashboardCoursesRequest) (*prot.DashboardCoursesResponse, error) {
	var courses []dto.DashboardCoursesResponse
	var total int64
	var err error

	// Nếu không truyền page và limit thì lấy hết
	if req.Page <= 0 && req.Limit <= 0 {
		// Lấy tất cả courses
		courses, err = s.dashboardCoursesRepo.GetDashboardCourses(req.SchoolId, req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}
		total = int64(len(courses))
	} else {
		// Set default values nếu có truyền page hoặc limit
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 10
		}

		// Get courses with pagination and search
		courses, err = s.dashboardCoursesRepo.GetDashboardCoursesWithPagination(req.SchoolId, req.Search, int(req.Page), int(req.Limit), req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}

		// Get total count for pagination
		total, err = s.dashboardCoursesRepo.GetDashboardCoursesCount(req.SchoolId, req.Search)
		if err != nil {
			return nil, err
		}
	}

	// Convert to protobuf format
	var protoCourses []*prot.DashboardReportCourse
	for _, course := range courses {
		// Convert TeacherInfos from DTO to protobuf
		var protoTeacherInfos []*prot.DashboardTeacherInfo
		for _, teacher := range course.TeacherInfos {
			protoTeacherInfos = append(protoTeacherInfos, &prot.DashboardTeacherInfo{
				Id:       teacher.ID,
				Username: teacher.Username,
				Name:     teacher.Name,
			})
		}

		protoCourse := &prot.DashboardReportCourse{
			Id:                                    course.ID,
			CourseId:                              course.CourseID,
			CourseName:                            course.CourseName,
			ObjectTitle:                           course.ObjectTitle,
			SchoolId:                              course.SchoolID,
			SchoolName:                            course.SchoolName,
			TotalStudents:                         course.TotalStudents,
			TotalTeachers:                         course.TotalTeachers,
			TeacherIds:                            course.TeacherIDs,
			TeacherInfos:                          protoTeacherInfos,
			ActiveTeachersFromSep8:                course.ActiveTeachersFromSep8,
			ActiveStudentsFromSep15:               course.ActiveStudentsFromSep15,
			StudentsCompletedHomeworkFromSep15:    course.StudentsCompletedHomeworkFromSep15,
			StudentsCompletedHomeworkSelectedWeek:     course.StudentsCompletedHomeworkSelectedWeek,
			ActiveStudentsSelectedWeek:                course.ActiveStudentsSelectedWeek,
			ActiveTeachersSelectedWeek:                course.ActiveTeachersSelectedWeek,
			TotalHomeworks:                        course.TotalHomeworks,
			AssignedHomeworks:                     course.AssignedHomeworks,
			CompletedHomeworks:                    course.CompletedHomeworks,
		}
		protoCourses = append(protoCourses, protoCourse)
	}

	// Calculate total pages
	var totalPages int32
	if req.Limit > 0 {
		totalPages = int32(math.Ceil(float64(total) / float64(req.Limit)))
	} else {
		totalPages = 1 // Nếu không có limit thì chỉ có 1 trang
	}

	return &prot.DashboardCoursesResponse{
		Courses:     protoCourses,
		Total:       total,
		Page:        req.Page,
		Limit:       req.Limit,
		TotalPages:  totalPages,
	}, nil
}
