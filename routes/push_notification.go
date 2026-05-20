package routes

import (
	"be-Clever School/controllers"
	"be-Clever School/middleware"
	"be-Clever School/repositories"
	"time"

	"github.com/gin-gonic/gin"
)

func RoutePushNotification(router *gin.Engine) {
	pushCtrl := controllers.NewPushNotificationController()
	authRepo := repositories.NewAuthRepository()

	api := router.Group("/api/v1/push")
	api.Use(middleware.TimeoutWithSkip(30*time.Second, []string{}))
	api.Use(middleware.AuthMiddleware(authRepo))
	{
		// Device management
		api.POST("/devices/register", pushCtrl.RegisterDevice)
		api.POST("/devices/unregister", pushCtrl.UnregisterDevice)
		api.GET("/devices", pushCtrl.GetMyDevices)

		// Test notification
		api.POST("/test", pushCtrl.SendTestNotification)

		// Notice push (admin only)
		api.POST("/notices/:id/send", pushCtrl.SendNoticeWithPush)
	}
}
