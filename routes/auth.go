package routes

import (
	"be-cleverschool/middleware"
	"be-cleverschool/redis"
	"be-cleverschool/repositories"
	"time"

	"github.com/gin-gonic/gin"
)

func RouteAuth(router *gin.Engine) {
	api := router.Group("/api")
	api.Use(middleware.TimeoutWithSkip(30*time.Second, []string{}))
	authController := NewAuthController()

	authRepo := repositories.NewAuthRepository()

	rateLimitLogin := redis.NewRateLimiter(5, time.Second, "login")
	rateLimitForgotPassword := redis.NewRateLimiter(3, time.Minute, "forgot_password")

	api.POST("/register", middleware.RateLimitMiddleware(rateLimitLogin), authController.Register)
	api.POST("/login", middleware.RateLimitMiddleware(rateLimitLogin), authController.Login)
	api.POST("/forgot-password", middleware.RateLimitMiddleware(rateLimitForgotPassword), authController.ForgotPassword)
	api.POST("/reset-password", authController.ResetPassword)
	api.GET("/sessions", middleware.AuthMiddleware(authRepo), authController.ListSessions)
	api.POST("/logout", middleware.AuthMiddleware(authRepo), authController.Logout)
	api.POST("/logout-all", middleware.AuthMiddleware(authRepo), authController.LogoutAll)
	api.GET("/profile", middleware.AuthMiddleware(authRepo), authController.Profile)
	api.PUT("/change-password", middleware.AuthMiddleware(authRepo), authController.ChangePassword)
	api.PUT("/accept-roles/:role", middleware.AuthMiddleware(authRepo), authController.AcceptRole)
}

