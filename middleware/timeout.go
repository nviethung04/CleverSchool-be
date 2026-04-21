package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns a middleware that limits handler execution time.
func Timeout(d time.Duration) gin.HandlerFunc {
	return TimeoutWithSkip(d, nil)
}

// TimeoutWithSkip returns a middleware that limits handler execution time,
// but skips timeout for paths in skipPaths (exact or prefix match).
// Use for long-running/streaming routes (SSE, WebSocket, long uploads).
func TimeoutWithSkip(d time.Duration, skipPaths []string) gin.HandlerFunc {
	skipMap := make(map[string]bool)
	for _, p := range skipPaths {
		if p != "" {
			skipMap[strings.TrimSuffix(p, "/")] = true
		}
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		pathTrimmed := strings.TrimSuffix(path, "/")
		if skipMap[pathTrimmed] {
			c.Next()
			return
		}
		for skipPath := range skipMap {
			if strings.HasPrefix(path, skipPath+"/") || path == skipPath {
				c.Next()
				return
			}
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		// Run the next handlers in a separate goroutine so we can stop waiting on timeout.
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			return
		case p := <-panicChan:
			panic(p)
		case <-ctx.Done():
			c.Writer.WriteHeader(http.StatusGatewayTimeout)
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"message": "request timed out",
			})
		}
	}
}
