package routes

import (
	"be-Clever School/controllers"
	"be-Clever School/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterMigrationRoutes(r *gin.Engine) {
	migrationController := new(controllers.MigrationController)
	api := r.Group("/api/upload")
	// Skip timeout cho /complete (xử lý async) và /extract (job dài), các route khác vẫn 60s
	api.Use(middleware.TimeoutWithSkip(60*time.Second, []string{"/api/upload/complete", "/api/upload/extract"}))
	migration := r.Group("/api/migration")
	{
		migration.POST("/upload-public", migrationController.UploadPublicDirectory)
	}

	uploadController := controllers.NewUploadS3Controller()

	api.POST("/presign", uploadController.PresignUpload)
	api.POST("/complete", uploadController.UploadComplete)
	api.POST("/extract", uploadController.Extract)
	api.GET("/progress", uploadController.GetExtractProgress)
}
