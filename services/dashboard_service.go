package services

import (
	"be-lms/prot"
	"be-lms/repositories"

	"github.com/gin-gonic/gin"
)

const (
	Month       = "month"
	Week        = "week"
	Day         = "day"
	Last24Hours = "24h"
	Up          = "up"
	Down        = "down"
	Same        = "same"
)

type DashboardService interface {
	DashboardAdmin(c *gin.Context) (*prot.DashboardAdminResponse, error)
	Export(c *gin.Context) (string, error)
}

type dashboardService struct {
	repo  repositories.DashboardRepository
	cache *DashboardCacheService
}

func NewDashboardService(repo repositories.DashboardRepository) DashboardService {
	return &dashboardService{
		repo:  repo,
		cache: NewDashboardCacheService(repo),
	}
}

func (s *dashboardService) DashboardAdmin(c *gin.Context) (*prot.DashboardAdminResponse, error) {
	tab := c.Query("tab")

	switch tab {
	case "overview":
		userRegister := s.cache.GetCachedUserRegister(c)
		studentRegister := s.cache.GetCachedStudentRegister(c)
		teacherRegister := s.cache.GetCachedTeacherRegister(c)
		activity := s.cache.GetCachedActivity(c)
		school := s.cache.GetCachedSchool(c)
		courseOverview := s.cache.GetCachedCourseOverview(c)

		return &prot.DashboardAdminResponse{
			UserOverview: &prot.UserOverview{
				User:     userRegister,
				Student:  studentRegister,
				Teacher:  teacherRegister,
				Activity: activity,
				School:   school,
			},
			CourseOverview:             courseOverview,
		}, nil
	case "quality":
		learningOverview  := s.cache.GetCachedLearningOverview(c)
		scoreDistribution := s.cache.GetCachedScoreDistribution(c)

		return &prot.DashboardAdminResponse{
			OverviewLearning:           learningOverview,
			ScoreDistributionOverview:  scoreDistribution,
		}, nil
	case "behavior":
		systemUsage := s.cache.GetCachedSystemUsage(c)
		return &prot.DashboardAdminResponse{
			SystemUsageOverview: systemUsage,
		}, nil
	case "grading":
		teacherPerformance := s.cache.GetCachedTeacherPerformance(c)
		return &prot.DashboardAdminResponse{
			TeacherPerformanceOverview: teacherPerformance,
		}, nil
	case "question":
		questionBank := s.cache.GetCachedQuestionBank(c)
		return &prot.DashboardAdminResponse{
			QuestionBankOverview:       questionBank,
		}, nil
	case "warning":
		riskWarning := s.cache.GetCachedRiskWarning(c)
		return &prot.DashboardAdminResponse{
			RiskAndWarning:             riskWarning,
		}, nil
	default:
		userRegister := s.cache.GetCachedUserRegister(c)
		studentRegister := s.cache.GetCachedStudentRegister(c)
		teacherRegister := s.cache.GetCachedTeacherRegister(c)
		activity := s.cache.GetCachedActivity(c)
		school := s.cache.GetCachedSchool(c)
		courseOverview := s.cache.GetCachedCourseOverview(c)
		learningOverview  := s.cache.GetCachedLearningOverview(c)
		scoreDistribution := s.cache.GetCachedScoreDistribution(c)
		systemUsage := s.cache.GetCachedSystemUsage(c)
		teacherPerformance := s.cache.GetCachedTeacherPerformance(c)
		questionBank := s.cache.GetCachedQuestionBank(c)
		riskWarning := s.cache.GetCachedRiskWarning(c)

		return &prot.DashboardAdminResponse{
			UserOverview: &prot.UserOverview{
				User:     userRegister,
				Student:  studentRegister,
				Teacher:  teacherRegister,
				Activity: activity,
				School:   school,
			},
			CourseOverview:             courseOverview,
			OverviewLearning:           learningOverview,
			ScoreDistributionOverview:  scoreDistribution,
			SystemUsageOverview:        systemUsage,
			TeacherPerformanceOverview: teacherPerformance,
			QuestionBankOverview:       questionBank,
			RiskAndWarning:             riskWarning,
		}, nil
	}
}
