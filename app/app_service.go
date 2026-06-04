package app

import (
	"be-lms/config"
	"be-lms/database/db"
	_ "be-lms/docs"
	"be-lms/i18n"
	"be-lms/jobs"
	"be-lms/middleware"
	"be-lms/redis"
	"be-lms/routes"
	"be-lms/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RunAppServer() {
	_ = godotenv.Load()

	utils.InitJWT()

	cfg := config.LoadConfig()
	config.InitDisks()
	config.InitLogger()

	config.Log.Warn("🚀 Ứng dụng đang khởi động...")
	config.Log.Warnf("🔧 App Debug Mode: %v", cfg.AppDebug)
	config.Log.Warnf("🌐 Port: %s", cfg.Port)

	// Set logging level based on LOG_LEVEL environment variable
	// This is separate from Gin's debug mode
	if cfg.AppDebug {
		// Only set Gin to debug mode if really needed for development
		// gin.SetMode(gin.DebugMode) // Comment out to avoid route logging
		gin.SetMode(gin.ReleaseMode) // Use release mode to avoid route logging
		config.Log.SetLevel(logrus.DebugLevel)
		config.Log.Debug("🐛 Running in DEBUG mode - Detailed logging enabled")
		config.Log.Debug("🔧 Gin set to Release mode to avoid route definition logging")
	} else {
		gin.SetMode(gin.ReleaseMode)
		config.Log.Warn("🚀 Running in RELEASE mode - Production logging enabled")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := db.ConnectPostgres(cfg); err != nil {
			config.Log.Error("Postgres error: %w", err)
			errChan <- fmt.Errorf("Postgres error: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := db.ConnectRedis(cfg); err != nil {
			config.Log.Warnf("Redis error: %s. Will continue without Redis.", err)
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		config.Log.Error("❌ Critical startup failure:", err)
		log.Fatal("❌ Critical startup failure:", err)
	}

	r := gin.Default()
	r.MaxMultipartMemory = 4 << 30

	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	sessionKey := strings.Replace(cfg.DBName, "_db_", "_", 1) + "_session"
	r.Use(sessions.Sessions(sessionKey, store))

	r.Use(middleware.CustomRecovery())

	// Set locale by header
	locale := "vi"
	i18n.Init(locale)
	r.Use(middleware.I18nMiddleware())

	// Set rate limit
	r = RateLimit(r)

	// Set cors by config
	r.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With",
			"Type", "Access-Control-Allow-Origin", "ngrok-skip-browser-warning", "Token",
			"Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset",
			"Upload-Defer-Length", "Upload-Concat", "Upload-Checksum", "Tus-Extension", "Tus-Max-Size",
			"Accept-Patch",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Type", "Date",
			"Location", "Tus-Resumable", "Tus-Version", "Tus-Extension", "Tus-Max-Size",
			"Upload-Offset", "Upload-Length", "Upload-Checksum",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Save activity log in database
	r.Use(middleware.ActivityLoggerMiddleware())

	routes.InitRoutes(r)
	if cfg.EnableSwagger {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Register observers
	RegisterObservers()

	// Add queue jobs
	go CronJob()

	// --- Setting timeout ---
	serverAddr := ":" + cfg.Port
	s := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  30 * time.Minute,
	}

	config.Log.Warnf("Server starting on %s", serverAddr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		config.Log.Fatal("❌ Main server failed to start:", err)
	}
}

func RegisterObservers() {
	db.RegisterGormCallbacks(db.MasterDB)
}

func RateLimit(r *gin.Engine) *gin.Engine {
	rateLimit := redis.NewRateLimiter(5000, time.Minute, "")
	r.Use(middleware.RateLimitMiddleware(rateLimit))
	return r
}

func CronJob() {
	jobs.StartCleanupCronJob()
	jobs.StartSyncMediaCronJob()
	jobs.StartHistoryUseCronJob()
	jobs.StartSyncKeywordCronJob()
	jobs.StartClearExportFilesCronJob()
	jobs.StartDatabaseBackupCronJob()
	jobs.StartHomeworkStatusScoringCronJob()

	// Dashboard background jobs (pre-cache + daily stats) — optional, not MVP-critical.
	if config.DashboardJobsEnabled() {
		jobs.StartDashboardCacheCronJob()
		jobs.StartDailySchoolStatisticsCronJob()
		jobs.StartDailyCourseStatisticsCronJob()
	} else {
		config.Log.Info("Dashboard background jobs disabled (set ENABLE_DASHBOARD_JOBS=true to enable)")
	}

	if config.S3Configured() {
		jobs.StartS3CleanupCronJob()
	}
}
