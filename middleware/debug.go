package middleware

import (
	"be-lms/config"
	"bytes"
	"io"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log chi tiết hơn
				stack := string(debug.Stack())
				stackLines := strings.Split(stack, "\n")

				// Log error
				config.Log.Errorf("Error: %v", err)
				config.Log.Errorf("Error Type: %T", err)

				// Log stack trace có cấu trúc
				config.Log.Errorf("Stack Trace:")
				for i, line := range stackLines {
					line = strings.TrimSpace(line)
					if line != "" {
						// Highlight các dòng quan trọng
						if strings.Contains(line, "be-lms") {
							config.Log.Errorf("%d: %s", i+1, line)
						} else {
							config.Log.Errorf("%d: %s", i+1, line)
						}
					}
				}

				// Log request body nếu có
				if c.Request.Body != nil {
					bodyBytes, _ := io.ReadAll(c.Request.Body)
					if len(bodyBytes) > 0 {
						// Giới hạn độ dài body để tránh log quá dài
						bodyStr := string(bodyBytes)
						if len(bodyStr) > 1000 {
							bodyStr = bodyStr[:1000] + "... [truncated]"
						}
					}
					// Restore body để có thể đọc lại
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}

				config.Log.Errorf("========================================")

				c.AbortWithStatus(500)
			}
		}()

		c.Next()
	}
}

// SafeGo wraps goroutine execution with panic recovery
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				stackLines := strings.Split(stack, "\n")

				config.Log.Errorf("========================================")
				config.Log.Errorf("Timestamp: %s", time.Now().Format(time.RFC3339))
				config.Log.Errorf("Error: %v", err)
				config.Log.Errorf("Error Type: %T", err)
				config.Log.Errorf("Stack Trace:")

				for i, line := range stackLines {
					line = strings.TrimSpace(line)
					if line != "" {
						if strings.Contains(line, "be-lms") {
							config.Log.Errorf("   %d: 🔥 %s", i+1, line)
						} else {
							config.Log.Errorf("   %d: %s", i+1, line)
						}
					}
				}

				config.Log.Errorf(" ========================================")
			}
		}()

		fn()
	}()
}
