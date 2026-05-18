package services

import (
	_ "be-Clever School/dto"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardListEntityService interface {
	GetSchools(c *gin.Context, req *requests.DashboardSchoolListRequest) (*prot.DashboardSchoolListResponse, error)
	GetCourses(c *gin.Context, req *requests.DashboardCourseListRequest) (*prot.DashboardCourseListResponse, error)
	GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) (*prot.DashboardTeacherListResponse, error)
	GetSubjects(c *gin.Context, req *requests.DashboardSubjectListRequest) (*prot.DashboardSubjectListResponse, error)
	GetExams(c *gin.Context, req *requests.DashboardExamListRequest) (*prot.DashboardExamListResponse, error)
	GetHomeworks(c *gin.Context, req *requests.DashboardHomeworkListRequest) (*prot.DashboardHomeworkListResponse, error)
	GetLessons(c *gin.Context, req *requests.DashboardLessonListRequest) (*prot.DashboardLessonListResponse, error)
	GetChapters(c *gin.Context, req *requests.DashboardChapterListRequest) (*prot.DashboardChapterListResponse, error)
	GetClasses(c *gin.Context, req *requests.DashboardClassListRequest) (*prot.DashboardClassListResponse, error)
	GetClassMains(c *gin.Context, req *requests.DashboardClassMainListRequest) (*prot.DashboardClassMainListResponse, error)
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
	roleID := utils.GetCurrentRoleId(c)
	userID := utils.GetCurrentUserId(c)

	onlyUserSchools := false
	if roleID == 2 && userID > 0 {
		onlyUserSchools = true
	}

	schools, totalCount, err := s.repo.GetSchools(c, int64(userID), onlyUserSchools, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardSchoolListResponse{
		Schools: resources.DashboardSchoolListCollection(schools),
		Total:   totalCount,
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

func (s *dashboardListEntityService) GetChapters(c *gin.Context, req *requests.DashboardChapterListRequest) (*prot.DashboardChapterListResponse, error) {
	chapters, totalCount, err := s.repo.GetChapters(req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardChapterListResponse{
		Chapters: resources.DashboardChapterListCollection(chapters),
		Total:    totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetClasses(c *gin.Context, req *requests.DashboardClassListRequest) (*prot.DashboardClassListResponse, error) {
	classes, totalCount, err := s.repo.GetClasses(c, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardClassListResponse{
		Classes: resources.DashboardClassListCollection(classes),
		Total:   totalCount,
	}, nil
}

func (s *dashboardListEntityService) GetClassMains(c *gin.Context, req *requests.DashboardClassMainListRequest) (*prot.DashboardClassMainListResponse, error) {
	classMains, totalCount, err := s.repo.GetClassMains(c, req)
	if err != nil {
		return nil, err
	}

	return &prot.DashboardClassMainListResponse{
		ClassMains: resources.DashboardClassMainListCollection(classMains),
		Total:      totalCount,
	}, nil
}
