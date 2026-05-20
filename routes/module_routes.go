package routes

import (
	"be-cleverschool/middleware"

	"github.com/gin-gonic/gin"
)

type ModuleController interface {
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type Restorer interface {
	Restore(c *gin.Context)
}

type Export interface {
	Export(c *gin.Context)
}

type Import interface {
	Import(c *gin.Context)
}

func RegisterModuleRoute(routerGroup *gin.RouterGroup, resource string, actions []string, controller ModuleController) {
	if contains(actions, "index") {
		routerGroup.GET("/"+resource, middleware.RoleMiddleware(resource+".index"), controller.GetAll)
	}
	if contains(actions, "show") {
		routerGroup.GET("/"+resource+"/:id", middleware.RoleMiddleware(resource+".show"), controller.GetByID)
	}
	if contains(actions, "store") {
		routerGroup.POST("/"+resource, middleware.RoleMiddleware(resource+".store"), controller.Create)
	}
	if contains(actions, "update") {
		routerGroup.PUT("/"+resource+"/:id", middleware.RoleMiddleware(resource+".update"), controller.Update)
	}
	if contains(actions, "destroy") {
		routerGroup.DELETE("/"+resource+"/:id", middleware.RoleMiddleware(resource+".destroy"), controller.Delete)
	}

	if contains(actions, "restore") {
		if restorer, ok := controller.(Restorer); ok {
			routerGroup.PATCH("/"+resource+"/:id/restore", middleware.RoleMiddleware(resource+".restore"), restorer.Restore)
		}
	}

	if contains(actions, "export") {
		if exporter, ok := controller.(Export); ok {
			routerGroup.GET("/"+resource+"/export", middleware.RoleMiddleware(resource+".export"), exporter.Export)
		}
	}

	if contains(actions, "import") {
		if importer, ok := controller.(Import); ok {
			routerGroup.POST("/"+resource+"/import", middleware.RoleMiddleware(resource+".import"), importer.Import)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

