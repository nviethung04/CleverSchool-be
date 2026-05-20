package services

import (
	"be-cleverschool/config"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"sync"

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
		var (
			userRegister    *prot.DashboardAdminItem
			studentRegister *prot.DashboardAdminItem
			teacherRegister *prot.DashboardAdminItem
			activity        *prot.DashboardAdminItem
			school          *prot.DashboardAdminItem
			userOnline      *prot.DashboardAdminItem
			courseOverview  *prot.CourseOverview
		)
		wg := &sync.WaitGroup{}
		wg.Add(7)
		go func() { defer wg.Done(); userRegister = s.cache.GetCachedUserRegister(c) }()
		go func() { defer wg.Done(); studentRegister = s.cache.GetCachedStudentRegister(c) }()
		go func() { defer wg.Done(); teacherRegister = s.cache.GetCachedTeacherRegister(c) }()
		go func() { defer wg.Done(); activity = s.cache.GetCachedActivity(c) }()
		go func() { defer wg.Done(); school = s.cache.GetCachedSchool(c) }()
		go func() { defer wg.Done(); userOnline = s.cache.GetCachedUserOnline(c) }()
		go func() { defer wg.Done(); courseOverview = s.cache.GetCachedCourseOverview(c) }()
		wg.Wait()

		return &prot.DashboardAdminResponse{
			UserOverview: &prot.UserOverview{
				User:       userRegister,
				Student:    studentRegister,
				Teacher:    teacherRegister,
				Activity:   activity,
				School:     school,
				UserOnline: userOnline,
			},
			CourseOverview: courseOverview,
		}, nil
	case "quality":
		learningOverview := s.cache.GetCachedLearningOverview(c)
		scoreDistribution := s.cache.GetCachedScoreDistribution(c)

		return &prot.DashboardAdminResponse{
			OverviewLearning:          learningOverview,
			ScoreDistributionOverview: scoreDistribution,
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
			QuestionBankOverview: questionBank,
		}, nil
	case "warning":
		riskWarning := s.cache.GetCachedRiskWarning(c)
		return &prot.DashboardAdminResponse{
			RiskAndWarning: riskWarning,
		}, nil
	default:
		var (
			userOverview       *prot.UserOverview
			userOverviewErr    error
			courseOverview     *prot.CourseOverview
			learningOverview   *prot.OverviewLearning
			scoreDistribution  *prot.ScoreDistributionOverview
			systemUsage        *prot.SystemUsageOverview
			teacherPerformance *prot.TeacherPerformanceOverview
			questionBank       *prot.QuestionBankOverview
			riskWarning        *prot.RiskAndWarning
		)
		wg := &sync.WaitGroup{}
		wg.Add(8)
		go func() { defer wg.Done(); userOverview, userOverviewErr = s.cache.GetCachedUserOverview(c) }()
		go func() { defer wg.Done(); courseOverview = s.cache.GetCachedCourseOverview(c) }()
		go func() { defer wg.Done(); learningOverview = s.cache.GetCachedLearningOverview(c) }()
		go func() { defer wg.Done(); scoreDistribution = s.cache.GetCachedScoreDistribution(c) }()
		go func() { defer wg.Done(); systemUsage = s.cache.GetCachedSystemUsage(c) }()
		go func() { defer wg.Done(); teacherPerformance = s.cache.GetCachedTeacherPerformance(c) }()
		go func() { defer wg.Done(); questionBank = s.cache.GetCachedQuestionBank(c) }()
		go func() { defer wg.Done(); riskWarning = s.cache.GetCachedRiskWarning(c) }()
		wg.Wait()

		if userOverviewErr != nil {
			config.Log.Error("GetCachedUserOverview error: ", userOverviewErr)
		}
		if userOverview != nil {
			config.Log.Info("UserOverview created successfully")
			if userOverview.UserOnline != nil {
				config.Log.Info("UserOnline data: ", userOverview.UserOnline)
			} else {
				config.Log.Warn("UserOnline is nil")
			}
		} else {
			config.Log.Warn("UserOverview is nil")
		}

		return &prot.DashboardAdminResponse{
			UserOverview:               userOverview,
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

