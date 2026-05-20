package routes

import (
	"be-Clever School/middleware"
	"be-Clever School/repositories"

	"github.com/gin-gonic/gin"
)

func RouteH5p(router *gin.Engine) {
	api := router.Group("/api")

	h5pGroup := api.Group("/h5p")
	authRepo := repositories.NewAuthRepository()

	h5pController := NewH5pController()

	h5pGroup.GET("/content", middleware.AuthMiddleware(authRepo), h5pController.ListContent)
	h5pGroup.GET("/content/:id", middleware.AuthMiddleware(authRepo), h5pController.ShowContent)
	h5pGroup.PUT("/content/:id", middleware.AuthMiddleware(authRepo), h5pController.UpdateContent)
	h5pGroup.DELETE("/content/:id", middleware.AuthMiddleware(authRepo), h5pController.DeleteContent)

	// Old

	h5pGroup.POST("/content-user-data", middleware.AuthMiddleware(authRepo), h5pController.StoreContentUserData)
	h5pGroup.POST("/content-score", middleware.AuthMiddleware(authRepo), h5pController.StoreScore)
	h5pGroup.GET("/content-data/:contentId/:userId", middleware.AuthMiddleware(authRepo), h5pController.GetContentUserData)
	h5pGroup.GET("/content-user-data/:contentId/:userId/:contextId", middleware.AuthMiddleware(authRepo), h5pController.GetContentUserDataByContentIdAndUser)

}
