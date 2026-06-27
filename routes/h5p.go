package routes

import (
	"be-lms/middleware"
	"be-lms/repositories"

	"github.com/gin-gonic/gin"
)

func RouteH5p(router *gin.Engine) {
	api := router.Group("/api")

	h5pGroup := api.Group("/h5p")
	authRepo := repositories.NewAuthRepository()

	h5pController := NewH5pController()

	h5pAuth := h5pGroup.Group("", middleware.AuthMiddleware(authRepo))
	h5pAuth.GET("/content", middleware.RoleMiddleware("h5p-contents.index"), h5pController.ListContent)
	h5pAuth.GET("/content/:id", middleware.RoleMiddleware("h5p-contents.show"), h5pController.ShowContent)
	h5pAuth.PUT("/content/:id", middleware.RoleMiddleware("h5p-contents.update"), h5pController.UpdateContent)
	h5pAuth.DELETE("/content/:id", middleware.RoleMiddleware("h5p-contents.destroy"), h5pController.DeleteContent)

	// Old

	h5pGroup.POST("/content-user-data", middleware.AuthMiddleware(authRepo), h5pController.StoreContentUserData)
	h5pGroup.POST("/content-score", middleware.AuthMiddleware(authRepo), h5pController.StoreScore)
	h5pGroup.GET("/content-data/:contentId/:userId", middleware.AuthMiddleware(authRepo), h5pController.GetContentUserData)
	h5pGroup.GET("/content-user-data/:contentId/:userId/:contextId", middleware.AuthMiddleware(authRepo), h5pController.GetContentUserDataByContentIdAndUser)

}
