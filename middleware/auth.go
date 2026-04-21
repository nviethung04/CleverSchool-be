package middleware

import (
	"be-lms/config"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/redis"
	"be-lms/repositories"
	"be-lms/utils"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AuthMiddleware(authRepo repositories.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Token")

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   i18n.Localize("messages.missing_token"),
				"message": i18n.Localize("messages.missing_token"),
			})
			return
		}

		claims, err := utils.CheckToken(c, authRepo, tokenStr)
		if err != nil {
			// Parse userID từ token (dù token hết hạn)
			userID, errId := utils.GetUserIDIgnoreExpiry(tokenStr)
			if errId == nil {
				// Lấy roleID từ token (không kiểm tra hạn)
				tokenObj, _, _ := new(jwt.Parser).ParseUnverified(tokenStr, &utils.CustomClaims{})
				customClaims, ok := tokenObj.Claims.(*utils.CustomClaims)
				if ok && customClaims.RoleID == models.StudentRoleId { // chỉ áp dụng cho học sinh
					latestToken, errLatest := redis.GetLatestUserToken(int(userID))
					if errLatest == nil && latestToken != "" && latestToken != tokenStr {
						config.Log.Infof("[TOKEN_KICKED] userID=%v, token=%v, latestToken=%v", userID, tokenStr, latestToken)
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error":   "Tài khoản của bạn đã được đăng nhập trên thiết bị khác!",
							"message": "Tài khoản của bạn đã được đăng nhập trên thiết bị khác!",
							"code":    "TOKEN_KICKED",
						})
						return
					}
				}
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   i18n.Localize("messages.invalid_token"),
				"message": i18n.Localize("messages.invalid_token"),
			})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("roleID", claims.RoleID)
		c.Set("schoolID", claims.SchoolID)
		c.Set("roleIDs", claims.RoleIDs)
		c.Set("roles", claims.Roles)
		c.Set("roleType", claims.RoleType)
		c.Set("memberType", claims.MemberType)

		c.Next()
	}
}

func RoleMiddleware(allowedPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var roleIDs []int

		roleType, _ := c.Get("roleType")

		// Ưu tiên lấy roleIDs
		if roleIDsVal, exists := c.Get("roleIDs"); exists {
			if ids, ok := roleIDsVal.([]int); ok {
				roleIDs = ids
			}
		}

		// Nếu chưa có, fallback sang roleID đơn
		if len(roleIDs) == 0 {
			if roleIDVal, exists := c.Get("roleID"); exists {
				if id, ok := roleIDVal.(int); ok {
					roleIDs = []int{id}
				}
			}
		}

		// Nếu vẫn không có -> token sai
		if len(roleIDs) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   i18n.Localize("messages.invalid_token"),
				"message": i18n.Localize("messages.invalid_token"),
			})
			c.Abort()
			return
		}

		if roleType != "" {
			filtered := make([]int, 0, len(roleIDs))
			for _, roleID := range roleIDs {
				switch roleType {
				case "admin":
					if roleID == models.AdminRoleId || roleID == models.SchoolRoleId || roleID == models.ReadOnlyRoleId {
						filtered = append(filtered, roleID)
					}
				case "teacher":
					if roleID == models.TeacherRoleId {
						filtered = append(filtered, roleID)
					}
				case "student":
					if roleID == models.StudentRoleId {
						filtered = append(filtered, roleID)
					}
				}
			}

			roleIDs = filtered
		}

		// Kiểm tra quyền
		hasPermission := false
		for _, roleID := range roleIDs {
			redis := redis.NewRoleRedis(roleID)
			permList, err := redis.GetPermissions()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   i18n.Localize("messages.invalid_token"),
					"message": i18n.Localize("messages.invalid_token"),
				})
				c.Abort()
				return
			}

			for _, perm := range permList {
				if perm == allowedPermission {
					hasPermission = true
					break
				}
			}

			if hasPermission {
				break
			}
		}

		// Cho qua hoặc chặn
		if hasPermission {
			c.Next()
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   i18n.Localize("messages.permission_denied"),
				"message": i18n.Localize("messages.permission_denied"),
			})
			c.Abort()
		}
	}
}

func RateLimitMiddleware(limiter *redis.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip rate limiting for WebSocket connections and public paths
		if strings.HasPrefix(path, "/public/") ||
			strings.HasPrefix(path, "/power-point/") ||
			strings.HasPrefix(path, "/ws/") { // Skip WebSocket connections
			c.Next()
			return
		}

		ip := c.ClientIP()

		isLimited, err := limiter.IsLimited(ip)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   i18n.Localize("messages.rate_limit_error"),
				"message": i18n.Localize("messages.rate_limit_error"),
				"Ip":      ip,
				"err":     err,
			})
			return
		}

		if isLimited {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   i18n.Localize("messages.rate_limit_error"),
				"message": i18n.Localize("messages.rate_limit_error"),
				"Ip":      ip,
			})
			return
		}

		c.Next()
	}
}

func BaseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cfg := config.LoadConfig()

		expectedUser := cfg.BasicAuthUsername
		expectedPass := cfg.BasicAuthPassword

		// inputPassHash := hashPassword(password)

		if username != expectedUser || expectedPass != password {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

func hashPassword(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}

func GetClientIP(c *gin.Context) string {
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback: RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		return ip
	}

	return ""
}
