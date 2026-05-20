package jobs

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

// DashboardCacheJob pre-caches dashboard data để người dùng có data ngay lập tức
func StartDashboardCacheCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Pre-cache dashboard data mỗi 30 phút (cùng với TTL của cache)
	c.AddFunc("*/30 * * * *", func() {
		PreCacheDashboardDataJob()
	})

	c.Start()
}

// PreCacheDashboardDataJob pre-cache tất cả dashboard data với các filter khác nhau
func PreCacheDashboardDataJob() {
	// Defer recover để tránh panic
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in PreCacheDashboardDataJob: %v", r)
		}
	}()

	config.Log.Info("Starting pre-cache dashboard data job")

	// Kiểm tra database connection
	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return
	}
	if db.ReplicaDB == nil {
		config.Log.Warn("ReplicaDB is nil - using MasterDB for all queries")
	}
	if db.RedisClient == nil {
		config.Log.Warn("RedisClient is nil - Redis not connected, cache operations will fail")
	}

	// Khởi tạo services
	dashboardRepo := repositories.NewDashboardRepository()
	dashboardCacheService := services.NewDashboardCacheService(dashboardRepo)

	func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Errorf("Panic in ClearAllDashboardCache: %v", r)
			}
		}()

		if err := dashboardCacheService.ClearAllDashboardCache(); err != nil {
			config.Log.Errorf("Error clearing dashboard cache: %v", err)
		}
	}()

	ctx := createContext()

	preCacheUserRegister(dashboardCacheService, ctx)
	preCacheStudentRegister(dashboardCacheService, ctx)
	preCacheTeacherRegister(dashboardCacheService, ctx)
	preCacheActivity(dashboardCacheService, ctx)
	preCacheSystemUsage(dashboardCacheService, ctx)
	preCacheCourseOverview(dashboardCacheService, ctx)
	preCacheLearningOverview(dashboardCacheService, ctx)
	preCacheScoreDistribution(dashboardCacheService, ctx)
	preCacheTeacherPerformance(dashboardCacheService, ctx)
	preCacheQuestionBank(dashboardCacheService, ctx)
	preCacheRiskWarning(dashboardCacheService, ctx)

	config.Log.Info("Finished pre-cache dashboard data job")
}

// Helper functions để pre-cache từng loại data
func preCacheUserRegister(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheUserRegister: %v", r)
		}
	}()

	dashboardCacheService.GetCachedUserRegister(ctx)
}

func preCacheStudentRegister(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheStudentRegister: %v", r)
		}
	}()

	dashboardCacheService.GetCachedStudentRegister(ctx)
}

func preCacheTeacherRegister(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheTeacherRegister: %v", r)
		}
	}()

	dashboardCacheService.GetCachedTeacherRegister(ctx)
}

func preCacheActivity(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheActivity: %v", r)
		}
	}()

	dashboardCacheService.GetCachedActivity(ctx)
}

func preCacheCourseOverview(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheCourseOverview: %v", r)
		}
	}()

	dashboardCacheService.GetCachedCourseOverview(ctx)
}

func preCacheLearningOverview(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheLearningOverview: %v", r)
		}
	}()

	dashboardCacheService.GetCachedLearningOverview(ctx)
}

func preCacheScoreDistribution(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheScoreDistribution: %v", r)
		}
	}()

	dashboardCacheService.GetCachedScoreDistribution(ctx)
}

func preCacheSystemUsage(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheSystemUsage: %v", r)
		}
	}()

	dashboardCacheService.GetCachedSystemUsage(ctx)
}

func preCacheTeacherPerformance(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheTeacherPerformance: %v", r)
		}
	}()

	dashboardCacheService.GetCachedTeacherPerformance(ctx)
}

func preCacheQuestionBank(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheQuestionBank: %v", r)
		}
	}()

	dashboardCacheService.GetCachedQuestionBank(ctx)
}

func preCacheRiskWarning(dashboardCacheService *services.DashboardCacheService, ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in preCacheRiskWarning: %v", r)
		}
	}()

	dashboardCacheService.GetCachedRiskWarning(ctx)
}

func createContext() *gin.Context {
	now := time.Now()

	yearFilter := strconv.Itoa(now.Year())
	monthFilter := strconv.Itoa(int(now.Month()))
	quarter := ((int(now.Month()) - 1) / 3) + 1
	quarterFilter := strconv.Itoa(quarter)

	// Tạo response recorder và context test
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Build query params
	q := url.Values{}
	q.Set("year", yearFilter)
	q.Set("quarter", quarterFilter)
	q.Set("month", monthFilter)

	// Fake request
	c.Request, _ = http.NewRequest("GET", "/?"+q.Encode(), nil)
	c.Request.Header.Set("X-Locale", "vi")

	return c
}

