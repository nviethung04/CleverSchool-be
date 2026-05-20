package routes

import (
	"be-Clever School/controllers"
	"be-Clever School/middleware"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func RouteMessage(router *gin.Engine) {
	api := router.Group("/api")
	authRepo := repositories.NewAuthRepository()

	chatMessageController := NewChatMessageController()
	chatReplyController := NewChatReplyController()

	managementRouter := api.Group("/manage")
	managementRouter.Use(middleware.AuthMiddleware(authRepo))

	// ===== WebSocket Chat =====
	// WebSocket endpoint - uses ticket-based auth (ticket validated in handler)
	// IMPORTANT: These routes must be registered BEFORE rate limiting and other middleware
	// to ensure WebSocket upgrade requests are not blocked
	chatWebSocketController := controllers.NewChatWebSocketController()
	
	// Generate WebSocket ticket endpoint (requires JWT in Authorization header)
	api.POST("/chat/ws-ticket", chatWebSocketController.GenerateWSTicket)
	
	// WebSocket connection endpoint (uses ticket from query param)
	router.GET("/ws/chat", chatWebSocketController.HandleWebSocket)
	
	// Health check endpoint to verify WebSocket route is accessible
	router.GET("/ws/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "websocket",
			"endpoint":  "/ws/chat",
			"timestamp": time.Now().Unix(),
			"message":   "WebSocket endpoint is reachable. Use /ws/chat for actual WebSocket connections.",
		})
	})
	
	// Test endpoint to verify WebSocket route is accessible
	router.GET("/ws/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "WebSocket endpoint is reachable. Use /ws/chat for actual WebSocket connections.",
		})
	})
	
	// Debug endpoint to test token validation (for development only)
	router.GET("/ws/debug-token", func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(400, gin.H{"error": "Missing token"})
			return
		}
		
		// Try to validate token
		claims, err := utils.CheckToken(c, authRepo, token)
		if err != nil {
			c.JSON(401, gin.H{
				"error": "Token validation failed",
				"details": err.Error(),
				"token_preview": func() string {
					if len(token) > 20 {
						return token[:20] + "..."
					}
					return token
				}(),
			})
			return
		}
		
		c.JSON(200, gin.H{
			"status": "ok",
			"user_id": claims.UserID,
			"role": claims.Role,
			"message": "Token is valid. WebSocket connection should work.",
		})
	})
	
	// WebSocket stats endpoints (require auth)
	managementRouter.GET("/ws/stats", chatWebSocketController.GetConnectionStats)
	managementRouter.GET("/courses/:id/ws/online-users", chatWebSocketController.GetOnlineUsers)

	managementRouter.POST("/courses/:id/chat/messages", chatMessageController.SendMessageWithMedias)
	managementRouter.GET("/courses/:id/chat/messages", chatMessageController.GetMessages)
	managementRouter.DELETE("/courses/:id/chat/messages/:messageId", chatMessageController.DeleteMessage)
	managementRouter.POST("/courses/:id/chat/messages/:messageId/pin", chatMessageController.TogglePinMessage)
	managementRouter.GET("/courses/:id/chat/messages/pinned", chatMessageController.GetPinnedMessages)
	managementRouter.GET("/courses/:id/chat/messages/count", chatMessageController.GetMessageCount)
	managementRouter.GET("/courses/:id/chat/messages/recent", chatMessageController.GetRecentMessages)

	managementRouter.POST("/courses/:id/chat/messages/:messageId/replies", chatReplyController.SendReply)
	// managementRouter.GET("/courses/:id/chat/messages/:messageId/replies", chatReplyController.GetReplies) // Tạm thời comment - chưa cần thiết
	managementRouter.GET("/courses/:id/chat/messages/:messageId/with-replies", chatReplyController.GetMessageWithReplies)

	// Private message APIs
	managementRouter.POST("/courses/:id/chat/messages/recipient/:recipientId", chatMessageController.SendToRecipientMessage)
	managementRouter.GET("/courses/:id/chat/messages/recipient/:recipientId", chatMessageController.GetFromRecipientMessage)
	managementRouter.GET("/courses/:id/chat/messages/recent-senders", chatMessageController.GetRecentSenders)
}
