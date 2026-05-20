package routes

import (
	"be-Clever School/middleware"
	"be-Clever School/repositories"
	"time"

	"github.com/gin-gonic/gin"
)

func RouteMedia(router *gin.Engine) {
	api := router.Group("/api")
	api.Use(middleware.TimeoutWithSkip(30*time.Second, []string{}))
	authRepo := repositories.NewAuthRepository()
	mediaController := NewMediaController()
	mediaGroup := api.Group("/", middleware.AuthMiddleware(authRepo))

	mediaGroup.GET("/medias/files/:folder_id", mediaController.Files)
	mediaGroup.GET("/medias/folders", mediaController.Folders)

	mediaGroup.Use(middleware.RoleMiddleware("medias.update"))
	{
		mediaGroup.POST("/upload-file", mediaController.UploadFile)
		mediaGroup.POST("/medias/folders", mediaController.UploadFolder)
		mediaGroup.DELETE("/medias/folders/:folder_id", mediaController.DeleteFolder)
		mediaGroup.DELETE("/medias/files/:file_id", mediaController.DeleteFile)
		mediaGroup.DELETE("/medias/folder-and-files/:folder_id", mediaController.DeleteFolderAndFiles)
	}
}
