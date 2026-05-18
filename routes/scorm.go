// routes/scorm.go
package routes

import (
	"be-Clever School/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterScormRoutes(r *gin.Engine) {
	// SCORM API endpoints
	scormController := controllers.NewScormController()
	scormUploadController := controllers.NewScormUploadController()

	// Launch SCORM activity
	r.GET("/api/scorm/launch/:id", scormController.LaunchScorm)

	// Get SCORM attempt details
	r.GET("/api/scorm/attempt/:id", scormController.GetScormAttempt)

	// List SCORM activities
	r.GET("/api/scorm/activities", scormController.ListScormActivities)

	// SCORM Package Upload endpoints
	r.GET("/scorm-upload", scormUploadController.UploadForm)
	r.POST("/api/scorm/upload", scormUploadController.UploadScormPackage)
	r.GET("/api/scorm/packages", scormUploadController.ListUploadedPackages)
	r.GET("/api/scorm/packages/:name", scormUploadController.GetPackageInfo)
	r.DELETE("/api/scorm/packages/:name", scormUploadController.DeletePackage)

	r.POST("/statements", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	// Serve SCORM runtime and wrapper
	r.Static("/static/scorm", "./static/scorm")

	// SCORM wrapper page
	r.GET("/scorm-viewer", func(c *gin.Context) {
		c.File("./static/scorm/scorm-wrapper.html")
	})

	// Serve static SCORM files (MUST be last to avoid conflicts)
	r.Static("/scorm", "./scorm-packages")

	// SCORM 1.2 API endpoints
	v12 := r.Group("/api/scorm/v12")
	{
		v12.GET("/get", scormController.GetScormData)
		v12.POST("/set", scormController.SetScormData)
		v12.POST("/commit", scormController.CommitScormData)
		v12.POST("/terminate", scormController.TerminateScormSession)
	}

	// SCORM 2004 API endpoints
	v2004 := r.Group("/api/scorm/v2004")
	{
		v2004.GET("/get", scormController.GetScormData)
		v2004.POST("/set", scormController.SetScormData)
		v2004.POST("/commit", scormController.CommitScormData)
		v2004.POST("/terminate", scormController.TerminateScormSession)
	}
}
