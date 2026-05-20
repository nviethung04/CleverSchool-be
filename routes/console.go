package routes

import (
	"be-Clever School/middleware"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterCliRoutes(r *gin.Engine) {
	console := r.Group("console")
	console.Use(middleware.BaseMiddleware())

	console.POST("/run-cli", func(c *gin.Context) {
		cli := c.Query("cli")
		if cli == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing cli parameter"})
			return
		}

		allowed := []string{
			"go run database/seeder/seeder.go",
			"go test ./command",
		}

		isAllowed := false
		for _, a := range allowed {
			if strings.Contains(cli, a) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized command"})
			return
		}

		args := []string{}
		args = strings.Fields(cli)

		cmd := exec.Command(args[0], args[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  err.Error(),
				"output": string(output),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Command executed successfully",
			"output":  string(output),
		})
	})
}
