package services

import (
	"be-lms/dto"
	"be-lms/prot"
	"be-lms/repositories"
	"math"
)

type DashboardCoursesProtoService interface {
	GetDashboardCoursesProto(req *prot.DashboardCoursesRequest) (*prot.DashboardCoursesResponse, error)
	GetCourseStudentsProto(req *prot.DashboardCourseStudentsRequest) (*prot.DashboardCourseStudentsResponse, error)
	GetCourseHomeworksProto(req *prot.DashboardCourseHomeworksRequest) (*prot.DashboardCourseHomeworksResponse, error)
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
		courses, err = s.dashboardCoursesRepo.GetDashboardCourses(req.SchoolId, req.TeacherId, req.CourseId, req.StartDate, req.EndDate)
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
		courses, err = s.dashboardCoursesRepo.GetDashboardCoursesWithPagination(req.SchoolId, req.TeacherId, req.CourseId, req.Search, int(req.Page), int(req.Limit), req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}

		// Get total count for pagination
		total, err = s.dashboardCoursesRepo.GetDashboardCoursesCount(req.SchoolId, req.TeacherId, req.CourseId, req.Search)
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
			Id:                            course.ID,
			CourseId:                      course.CourseID,
			CourseName:                    course.CourseName,
			ObjectTitle:                   course.ObjectTitle,
			SchoolId:                      course.SchoolID,
			SchoolName:                    course.SchoolName,
			TotalStudents:                 course.TotalStudents,
			TotalTeachers:                 course.TotalTeachers,
			TeacherIds:                    course.TeacherIDs,
			TeacherInfos:                  protoTeacherInfos,
			ActiveTeachers:                course.ActiveTeachersSelectedWeek,
			ActiveStudents:                course.ActiveStudentsSelectedWeek,
			StudentsCompletedHomework:     course.StudentsCompletedHomeworkSelectedWeek,
			TotalHomeworks:                course.TotalHomeworks,
			AssignedHomeworks:             course.AssignedHomeworks,
			StudentsCompletedAllHomeworks: course.StudentsCompletedAllHomeworks,
			StudentsDoingHomeworks:        course.StudentsDoingHomeworks,
			StudentsNotStartedAnyHomework: course.StudentsNotStartedAnyHomework,
			StudentActiveNotStartedAnyHomework: course.StudentActiveNotStartedAnyHomework,
			HomeworkOver_50PercentStudentComplete: course.HomeworkOver50PercentStudentComplete,
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

// GetCourseStudentsProto lấy danh sách học sinh trong course kèm thông tin homework
func (s *dashboardCoursesProtoService) GetCourseStudentsProto(req *prot.DashboardCourseStudentsRequest) (*prot.DashboardCourseStudentsResponse, error) {
	// Gọi repository để lấy dữ liệu
	response, err := s.dashboardCoursesRepo.GetCourseStudents(req.CourseId, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	// Convert DTO sang protobuf
	var protoStudents []*prot.DashboardCourseStudent
	for _, student := range response.Students {
		var protoHomeworks []*prot.DashboardCourseStudentHomework
		for _, homework := range student.Homeworks {
			protoHomeworks = append(protoHomeworks, &prot.DashboardCourseStudentHomework{
				Id:              homework.ID,
				Name:            homework.Name,
				IsCompleted:     homework.IsCompleted,
				IsCompletedLate: homework.IsCompletedLate,
			})
		}

		protoStudent := &prot.DashboardCourseStudent{
			Id:        student.ID,
			Name:      student.Name,
			IsActive:  student.IsActive,
			Homeworks: protoHomeworks,
		}
		protoStudents = append(protoStudents, protoStudent)
	}

	return &prot.DashboardCourseStudentsResponse{
		Students: protoStudents,
	}, nil
}

func (s *dashboardCoursesProtoService) GetCourseHomeworksProto(req *prot.DashboardCourseHomeworksRequest) (*prot.DashboardCourseHomeworksResponse, error) {
	response, err := s.dashboardCoursesRepo.GetCourseHomeworks(req.CourseId, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	var protoHomeworks []*prot.DashboardCourseHomework
	for _, homework := range response.Homeworks {
		var assignedAt int64
		if homework.AssignedAt != nil {
			assignedAt = homework.AssignedAt.Unix()
		}
		protoHomeworks = append(protoHomeworks, &prot.DashboardCourseHomework{
			Id:             homework.ID,
			Name:           homework.Name,
			AssignedAt:     assignedAt,
			IsAssignedLate: homework.IsAssignedLate,
		})
	}

	return &prot.DashboardCourseHomeworksResponse{
		Homeworks: protoHomeworks,
	}, nil
}
