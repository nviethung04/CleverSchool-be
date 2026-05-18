package routes

import (
	"be-Clever School/controllers"
	"be-Clever School/middleware"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"time"

	"github.com/gin-gonic/gin"
)

func RouteDashboard(router *gin.Engine) {
	api := router.Group("/api")

	api.Use(middleware.TimeoutWithSkip(60*time.Second, []string{}))

	authRepo := repositories.NewAuthRepository()

	dashboardRepository := repositories.NewDashboardRepository()
	dashboardService := services.NewDashboardService(dashboardRepository)
	dashboardController := controllers.NewDashboardController(dashboardService)

	api.GET("/dashboard", middleware.AuthMiddleware(authRepo), dashboardController.Dashboard)
	api.GET("/dashboard/export", middleware.AuthMiddleware(authRepo), dashboardController.Export)

	// Dashboard Report
	dashboardReportController := controllers.NewDashboardReportController()
	api.GET("/dashboard/report", middleware.AuthMiddleware(authRepo), dashboardReportController.GetDashboardReport)

	// Dashboard Schools
	dashboardSchoolsRepository := repositories.NewDashboardSchoolsRepository()
	dashboardSchoolsService := services.NewDashboardSchoolsService(dashboardSchoolsRepository)
	dashboardSchoolsController := controllers.NewDashboardSchoolsController(dashboardSchoolsService)
	api.GET("/dashboard/report/schools/export", middleware.AuthMiddleware(authRepo), dashboardSchoolsController.ExportDashboardSchools)

	// Dashboard Schools Proto (thay thế API cũ)
	dashboardSchoolsProtoService := services.NewDashboardSchoolsProtoService(dashboardSchoolsRepository)
	dashboardSchoolsProtoController := controllers.NewDashboardSchoolsProtoController(dashboardSchoolsProtoService)
	api.GET("/dashboard/report/schools", middleware.AuthMiddleware(authRepo), dashboardSchoolsProtoController.GetDashboardSchoolsProto)

	// Dashboard Courses
	dashboardCoursesRepository := repositories.NewDashboardCoursesRepository()
	dashboardCoursesService := services.NewDashboardCoursesService(dashboardCoursesRepository)
	dashboardCoursesController := controllers.NewDashboardCoursesController(dashboardCoursesService)
	api.GET("/dashboard/report/courses/export", middleware.AuthMiddleware(authRepo), dashboardCoursesController.ExportDashboardCoursesAll)

	// Dashboard Courses Proto (thay thế API cũ)
	dashboardCoursesProtoService := services.NewDashboardCoursesProtoService(dashboardCoursesRepository)
	dashboardCoursesProtoController := controllers.NewDashboardCoursesProtoController(dashboardCoursesProtoService)
	api.GET("/dashboard/report/courses", middleware.AuthMiddleware(authRepo), dashboardCoursesProtoController.GetDashboardCoursesProto)
	api.GET("/dashboard/report/courses/students", middleware.AuthMiddleware(authRepo), dashboardCoursesProtoController.GetCourseStudentsProto)
	api.GET("/dashboard/report/courses/homeworks", middleware.AuthMiddleware(authRepo), dashboardCoursesProtoController.GetCourseHomeworksProto)

}
