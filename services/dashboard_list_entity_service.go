package services

import (
	_ "be-lms/dto"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardListEntityService interface {
	GetSchools(c *gin.Context, req *requests.DashboardSchoolListRequest) (*prot.DashboardSchoolListResponse, error)
	GetPrograms(c *gin.Context, req *requests.DashboardProgramListRequest) (*prot.DashboardProgramListResponse, error)
	GetCourses(c *gin.Context, req *requests.DashboardCourseListRequest) (*prot.DashboardCourseListResponse, error)
	GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) (*prot.DashboardTeacherListResponse, error)
	GetSubjects(c *gin.Context, req *requests.DashboardSubjectListRequest) (*prot.DashboardSubjectListResponse, error)
	GetExams(c *gin.Context, req *requests.DashboardExamListRequest) (*prot.DashboardExamListResponse, error)
	GetHomeworks(c *gin.Context, req *requests.DashboardHomeworkListRequest) (*prot.DashboardHomeworkListResponse, error)
	GetExercises(c *gin.Context, req *requests.DashboardHomeworkListRequest) (map[string]interface{}, error)
	GetLessons(c *gin.Context, req *requests.DashboardLessonListRequest) (*prot.DashboardLessonListResponse, error)
}

type dashboardListEntityService struct {
	repo repositories.DashboardListEntityRepository
}

func NewDashboardListEntityService() DashboardListEntityService {
	return &dashboardListEntityService{
		repo: repositories.NewDashboardListEntityRepository(),
	}
}

func (s *dashboardListEntityService) GetSchools(c *gin.Context, req *requests.DashboardSchoolListRequest) (*prot.DashboardSchoolListResponse, error) {
	schools, totalCount, err := s.repo.GetSchools(c, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardSchoolListResponse{
		Schools: resources.DashboardSchoolListCollection(schools),
		Total:   totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetPrograms(c *gin.Context, req *requests.DashboardProgramListRequest) (*prot.DashboardProgramListResponse, error) {
	programs, totalCount, err := s.repo.GetPrograms(c, req.SchoolID, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardProgramListResponse{
		Programs: resources.DashboardProgramListCollection(programs),
		Total:    totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetCourses(c *gin.Context, req *requests.DashboardCourseListRequest) (*prot.DashboardCourseListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserCourses := false
	if roleID != 1 && userID > 0 {
		onlyUserCourses = true
	}

	courses, totalCount, err := s.repo.GetCourses(c, int64(userID), req.SchoolID, onlyUserCourses, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardCourseListResponse{
		Courses: resources.DashboardCourseListCollection(courses),
		Total:   totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) (*prot.DashboardTeacherListResponse, error) {
	teachers, totalCount, err := s.repo.GetTeachers(c, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardTeacherListResponse{
		Teachers: resources.DashboardTeacherListCollection(teachers),
		Total:    totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetSubjects(c *gin.Context, req *requests.DashboardSubjectListRequest) (*prot.DashboardSubjectListResponse, error) {
	subjects, totalCount, err := s.repo.GetSubjects(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardSubjectListResponse{
		Subjects: resources.DashboardSubjectListCollection(subjects),
		Total:    totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetExams(c *gin.Context, req *requests.DashboardExamListRequest) (*prot.DashboardExamListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserExams := false
	if roleID == 2 && userID > 0 {
		onlyUserExams = true
	}

	exams, totalCount, err := s.repo.GetExams(int64(userID), onlyUserExams, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardExamListResponse{
		Exams: resources.DashboardExamListCollection(exams),
		Total: totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetHomeworks(c *gin.Context, req *requests.DashboardHomeworkListRequest) (*prot.DashboardHomeworkListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserHomeworks := false
	if roleID == 2 && userID > 0 {
		onlyUserHomeworks = true
	}

	homeworks, totalCount, err := s.repo.GetHomeworks(int64(userID), onlyUserHomeworks, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardHomeworkListResponse{
		Homeworks: resources.DashboardHomeworkListCollection(homeworks),
		Total:     totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetExercises(c *gin.Context, req *requests.DashboardHomeworkListRequest) (map[string]interface{}, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserExercises := false
	if roleID == 2 && userID > 0 {
		onlyUserExercises = true
	}

	exercises, totalCount, err := s.repo.GetExercises(int64(userID), onlyUserExercises, req)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(exercises))
	for _, e := range exercises {
		item := map[string]interface{}{
			"id":           strconv.FormatInt(e.ID, 10),
			"name":         e.Name,
			"course_id":    e.CourseID,
			"subject_id":   e.SubjectID,
			"course_name":  e.CourseName,
			"subject_name": e.SubjectName,
			"lesson_id":    e.LessonID,
			"lesson_title": e.LessonTitle,
		}
		items = append(items, item)
	}

	return map[string]interface{}{
		"exercises": items,
		"total":     totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetLessons(c *gin.Context, req *requests.DashboardLessonListRequest) (*prot.DashboardLessonListResponse, error) {
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserLessons := false
	if roleID == 2 && userID > 0 {
		onlyUserLessons = true
	}

	lessons, totalCount, err := s.repo.GetLessons(int64(userID), onlyUserLessons, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardLessonListResponse{
		Lessons: resources.DashboardLessonListCollection(lessons),
		Total:   totalCount,
	}, nil
}
