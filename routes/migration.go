package routes

import (
	"be-lms/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterMigrationRoutes(r *gin.Engine) {
	migrationController := new(controllers.MigrationController)
	migration := r.Group("/api/migration")
	{
		migration.POST("/upload-public", migrationController.UploadPublicDirectory)
	}

	uploadController := controllers.NewUploadS3Controller()

	r.POST("/api/upload/presign", uploadController.PresignUpload)
	r.POST("/api/upload/complete", uploadController.UploadComplete)
	r.POST("/api/upload/extract", uploadController.Extract)
}
