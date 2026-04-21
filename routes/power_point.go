package routes

import (
	"be-lms/controllers"
	"be-lms/services"

	"github.com/gin-gonic/gin"
)

func RoutePowerPoint(router *gin.Engine) {
	powerPointGroup := router.Group("/power-point")

	powerPointService := services.NewPowerPointService()
	powerPointController := controllers.NewPowerPointController(powerPointService)

	router.PUT("api/power-point/scan-by-path/:path", powerPointController.ScanByPath)

	// Single catch-all for nested folders/packages and assets
	powerPointGroup.GET("/*any", powerPointController.ServeDynamicPowerPoint)
}
