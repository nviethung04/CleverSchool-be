package controllers

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/redis"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	ws "be-Clever School/websocket"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"be-Clever School/database/db"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// getUpgrader returns a WebSocket upgrader with proper origin checking
func getUpgrader() websocket.Upgrader {
	cfg := config.LoadConfig()
	allowedOrigins := cfg.AllowOrigins

	return websocket.Upgrader{
		ReadBufferSize:  4096, // Increased for better performance
		WriteBufferSize: 4096, // Increased for better performance
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			host := r.Host
			requestURL := r.URL.String()

			// Debug log with more headers for troubleshooting
			config.Log.Infof("[WS CheckOrigin] Origin: %s, Host: %s, URL: %s, X-Forwarded-For: %s, X-Forwarded-Proto: %s, CF-Ray: %s, CF-Connecting-IP: %s",
				origin, host, requestURL, r.Header.Get("X-Forwarded-For"), r.Header.Get("X-Forwarded-Proto"),
				r.Header.Get("CF-Ray"), r.Header.Get("CF-Connecting-IP"))
			config.Log.Infof("[WS CheckOrigin] Allowed origins from config: %v", allowedOrigins)

			// If no origin header, allow (for same-origin requests or testing)
			// This can happen when connecting from same origin or when browser doesn't send Origin header
			if origin == "" {
				config.Log.Debugf("[WS CheckOrigin] No origin header, allowing connection (same-origin or direct connection)")
				return true
			}

			// If no allowed origins configured, allow all (for development)
			if len(allowedOrigins) == 0 {
				config.Log.Warnf("[WS CheckOrigin] No allowed origins configured, allowing all - origin=%s", origin)
				return true
			}

			// Normalize origin (remove trailing slash)
			origin = strings.TrimSuffix(origin, "/")

			// Check if origin is in allowed list
			for _, allowedOrigin := range allowedOrigins {
				// Normalize allowed origin
				allowedOrigin = strings.TrimSpace(allowedOrigin)
				allowedOrigin = strings.TrimSuffix(allowedOrigin, "/")

				// Exact match
				if origin == allowedOrigin {
					config.Log.Infof("[WS CheckOrigin] ✅ Origin allowed (exact match): %s", origin)
					return true
				}

				// Support domain matching (e.g., allow https://Clever Schoolx-pre-api-chat.xClever School.vn and http://Clever Schoolx-pre-api-chat.xClever School.vn)
				// Extract domain from allowed origin
				allowedDomain := extractDomain(allowedOrigin)
				requestDomain := extractDomain(origin)

				if allowedDomain != "" && requestDomain != "" && allowedDomain == requestDomain {
					config.Log.Infof("[WS CheckOrigin] ✅ Origin allowed (domain match): %s matches %s (domain: %s)", origin, allowedOrigin, allowedDomain)
					return true
				}

				// Also check if the request is coming from the same domain as the API
				// This handles cases where frontend and API are on different subdomains but same base domain
				// e.g., frontend: https://Clever Schoolx-pre.xClever School.vn, API: https://Clever Schoolx-pre-api-chat.xClever School.vn
				if allowedDomain != "" && requestDomain != "" {
					// Check if they share the same base domain (e.g., both end with .xClever School.vn)
					allowedBaseDomain := getBaseDomain(allowedDomain)
					requestBaseDomain := getBaseDomain(requestDomain)

					if allowedBaseDomain != "" && requestBaseDomain != "" && allowedBaseDomain == requestBaseDomain {
						config.Log.Infof("[WS CheckOrigin] ✅ Origin allowed (base domain match): %s shares base domain with %s (base: %s)",
							origin, allowedOrigin, allowedBaseDomain)
						return true
					}
				}
			}

			config.Log.Warnf("[WS CheckOrigin] ❌ Origin rejected: %s (not in allowed list: %v)", origin, allowedOrigins)
			config.Log.Warnf("[WS CheckOrigin] Request details - Host: %s, URL: %s", host, requestURL)
			return false
		},
		// Enable compression for better performance
		EnableCompression: true,
		// WORKAROUND: Don't check for required headers - let gorilla handle it
		// This helps with Traefik proxy issues
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			config.Log.Errorf("[WS Error] Upgrade error - Status: %d, Reason: %v", status, reason)
			config.Log.Errorf("[WS Error] Request headers: %+v", r.Header)
			http.Error(w, reason.Error(), status)
		},
	}
}

// extractDomain extracts the domain from a URL (e.g., "https://Clever Schoolx-pre-api-chat.xClever School.vn" -> "Clever Schoolx-pre-api-chat.xClever School.vn")
func extractDomain(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	// Remove protocol
	if idx := strings.Index(urlStr, "://"); idx != -1 {
		urlStr = urlStr[idx+3:]
	}

	// Remove path and port
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	if idx := strings.Index(urlStr, ":"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	return urlStr
}

// getBaseDomain extracts the base domain (e.g., "Clever Schoolx-pre-api-chat.xClever School.vn" -> "xClever School.vn")
// This helps match subdomains of the same base domain
func getBaseDomain(domain string) string {
	if domain == "" {
		return ""
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}

	// Return last 2 parts (e.g., "xClever School.vn")
	// For domains like "example.co.uk", this would return "co.uk"
	// For most cases like "xClever School.vn", it returns "xClever School.vn"
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}

	return domain
}

// ChatWebSocketController handles WebSocket connections for chat
type ChatWebSocketController struct {
	hub           *ws.Hub
	chatService   services.ChatMessageService
	replyService  services.ChatReplyService
	pubsubService services.ChatPubSubService
	chatResource  resources.ChatResource
	authRepo      repositories.AuthRepository

	// Track subscribed courses to avoid duplicate Redis subscriptions
	subscribedCourses map[uint64]bool
	subscribeMu       sync.Mutex
}

// NewChatWebSocketController creates a new ChatWebSocketController
func NewChatWebSocketController() *ChatWebSocketController {
	// Initialize repositories
	chatMessageRepo := repositories.NewChatMessageRepository()
	userRepo := repositories.NewUserRepository()
	courseRepo := repositories.NewCourseRepository()
	messageMediaRepo := repositories.NewMessageMediaRepository()
	mediaRepo := repositories.NewMediaRepository()

	return &ChatWebSocketController{
		hub: ws.GetHub(),
		chatService: services.NewChatMessageService(
			chatMessageRepo,
			userRepo,
			courseRepo,
			messageMediaRepo,
			mediaRepo,
		),
		replyService: services.NewChatReplyService(
			chatMessageRepo,
			userRepo,
			courseRepo,
		),
		pubsubService:     services.NewChatPubSubService(),
		chatResource:      resources.NewChatResource(),
		authRepo:          repositories.NewAuthRepository(),
		subscribedCourses: make(map[uint64]bool),
	}
}

// WSTicketInfo stores ticket information in Redis
type WSTicketInfo struct {
	UserID    uint64 `json:"user_id"`
	RoleID    int    `json:"role_id"`
	Role      string `json:"role"`
	CourseID  uint64 `json:"course_id,omitempty"` // Optional, can be set when ticket is used
	CreatedAt int64  `json:"created_at"`
}

// GenerateWSTicket generates a short-lived ticket for WebSocket connection
// @Summary Generate WebSocket Ticket
// @Description Generate a short-lived ticket (60s) for WebSocket connection using JWT token
// @Tags Chat WebSocket
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT Token"
// @Param course_id query int false "Course ID (optional, can be set later)"
// @Success 200 {object} map[string]interface{} "Ticket generated successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/chat/ws-ticket [post]
func (c *ChatWebSocketController) GenerateWSTicket(ctx *gin.Context) {
	// Get JWT token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
		return
	}

	// Extract token from "Bearer <token>"
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return
	}

	// Validate JWT token
	claims, err := utils.CheckTokenForWebSocket(tokenStr)
	if err != nil {
		config.Log.Warnf("WSTicket generation failed: invalid token (IP: %s, error: %v)", ctx.ClientIP(), err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
		return
	}

	// Get optional course_id from query
	var courseID uint64
	if courseIDStr := ctx.Query("course_id"); courseIDStr != "" {
		courseID, err = strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			config.Log.Warnf("WSTicket generation: invalid course_id=%s (IP: %s)", courseIDStr, ctx.ClientIP())
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id"})
			return
		}
	}

	// Generate random ticket (16 bytes = 24 chars base64)
	ticketBytes := make([]byte, 16)
	if _, err := rand.Read(ticketBytes); err != nil {
		config.Log.Errorf("WSTicket generation: failed to generate random bytes (IP: %s, error: %v)", ctx.ClientIP(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate ticket"})
		return
	}
	ticket := base64.URLEncoding.EncodeToString(ticketBytes)

	// Store ticket in Redis with 60s TTL
	ticketInfo := WSTicketInfo{
		UserID:    uint64(claims.UserID),
		RoleID:    claims.RoleID,
		Role:      claims.Role,
		CourseID:  courseID,
		CreatedAt: time.Now().Unix(),
	}

	ticketData, err := json.Marshal(ticketInfo)
	if err != nil {
		config.Log.Errorf("WSTicket generation: failed to marshal ticket info (IP: %s, error: %v)", ctx.ClientIP(), err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	ticketKey := "ws_ticket:" + ticket
	if db.RedisClient != nil {
		err = db.RedisClient.Set(db.Ctx, ticketKey, ticketData, 60*time.Second).Err()
		if err != nil {
			config.Log.Errorf("WSTicket generation: failed to store ticket in Redis (IP: %s, error: %v)", ctx.ClientIP(), err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store ticket"})
			return
		}
	} else {
		config.Log.Warnf("WSTicket generation: Redis not available, ticket will not be stored")
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not available"})
		return
	}

	config.Log.Infof("WSTicket generated: user_id=%d, course_id=%d, ticket=%s", claims.UserID, courseID, ticket[:8]+"...")

	ctx.JSON(http.StatusOK, gin.H{
		"ticket":     ticket,
		"expires_in": 60,
		"message":    "Ticket generated successfully. Use this ticket to connect to WebSocket within 60 seconds.",
	})
}

// HandleWebSocket handles WebSocket upgrade and connection
// @Summary WebSocket Chat Connection
// @Description Upgrade HTTP connection to WebSocket for real-time chat using ticket (preferred) or token (legacy)
// @Tags Chat WebSocket
// @Param course_id query string true "Course ID"
// @Param ticket query string false "WebSocket Ticket (from /api/chat/ws-ticket) - PREFERRED"
// @Param token query string false "JWT Token (LEGACY - not recommended, may be blocked by WAF)"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /ws/chat [get]
func (c *ChatWebSocketController) HandleWebSocket(ctx *gin.Context) {
	// Get parameters from query string
	ticketStr := ctx.Query("ticket")
	tokenStr := ctx.Query("token") // Legacy support
	courseIDStr := ctx.Query("course_id")

	// Log incoming request for debugging
	config.Log.Infof("WebSocket connection request: IP=%s, Origin=%s, User-Agent=%s, Path=%s",
		ctx.ClientIP(), ctx.GetHeader("Origin"), ctx.GetHeader("User-Agent"), ctx.Request.URL.Path)

	// Support both ticket (preferred) and token (legacy) for backward compatibility
	var userID uint64
	var courseID uint64
	var err error

	if ticketStr != "" {
		// Ticket-based authentication - not implemented, fallback to token
		config.Log.Warnf("WebSocket ticket authentication requested but not implemented, falling back to token")
		tokenStr = ticketStr // Use ticket as token for now (backward compatibility)
	}

	if tokenStr != "" {
		// LEGACY METHOD: Token-based authentication (for backward compatibility)
		config.Log.Warnf("WebSocket using legacy token-based authentication (not recommended - may be blocked by WAF)")

		// Validate course_id
		if courseIDStr == "" {
			config.Log.Warnf("WebSocket connection rejected: missing course_id (IP: %s)", ctx.ClientIP())
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing course_id parameter"})
			return
		}

		courseID, err = strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			config.Log.Warnf("WebSocket connection rejected: invalid course_id=%s (IP: %s, error: %v)", courseIDStr, ctx.ClientIP(), err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course_id"})
			return
		}

		// Validate JWT token (stateless validation for WebSocket - no session store check)
		claims, err := utils.CheckTokenForWebSocket(tokenStr)
		if err != nil {
			config.Log.Warnf("WebSocket connection rejected: invalid token (IP: %s, course_id: %d, error: %v)", ctx.ClientIP(), courseID, err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		userID = uint64(claims.UserID)
	} else {
		// Neither ticket nor token provided
		config.Log.Warnf("WebSocket connection rejected: missing ticket or token (IP: %s)", ctx.ClientIP())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":           "Missing ticket or token parameter",
			"message":         "Please use ticket (preferred) or token (legacy) for authentication",
			"ticket_endpoint": "/api/chat/ws-ticket",
		})
		return
	}

	// Continue with WebSocket upgrade (userID and courseID already validated)
	config.Log.Infof("WebSocket connection attempt: user_id=%d, course_id=%d, IP=%s, Origin=%s", userID, courseID, ctx.ClientIP(), ctx.GetHeader("Origin"))

	// Debug: Log all headers for troubleshooting
	config.Log.Infof("[WS Debug] Headers - Upgrade: %s, Connection: %s, Sec-WebSocket-Key: %s, Sec-WebSocket-Version: %s",
		ctx.GetHeader("Upgrade"), ctx.GetHeader("Connection"), ctx.GetHeader("Sec-WebSocket-Key"), ctx.GetHeader("Sec-WebSocket-Version"))

	// Upgrade to WebSocket
	upgrader := getUpgrader()

	// WORKAROUND: Handle Traefik proxy that strips WebSocket headers
	// If Connection/Upgrade headers are missing but Sec-Websocket-Key is present, it's a WebSocket request
	connection := ctx.GetHeader("Connection")
	upgrade := ctx.GetHeader("Upgrade")
	secWebsocketKey := ctx.GetHeader("Sec-WebSocket-Key")

	if secWebsocketKey != "" && (connection == "" || upgrade == "") {
		config.Log.Warnf("[WS Workaround] Traefik stripped WebSocket headers, manually setting them")
		// Traefik removed headers, add them back manually for gorilla websocket
		ctx.Request.Header.Set("Connection", "Upgrade")
		ctx.Request.Header.Set("Upgrade", "websocket")
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		config.Log.Errorf("WebSocket upgrade failed: user_id=%d, course_id=%d, IP=%s, error=%v", userID, courseID, ctx.ClientIP(), err)
		config.Log.Errorf("[WS Debug] Full request headers: %+v", ctx.Request.Header)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed", "details": err.Error()})
		return
	}

	config.Log.Infof("WebSocket connection established: user_id=%d, course_id=%d", userID, courseID)

	// Create safe connection wrapper
	safeConn := ws.NewSafeConn(conn)

	// Create client using NewClient to ensure all channels are initialized
	client := ws.NewClient(c.hub, safeConn, userID, courseID)

	// Start write pump FIRST (so it can receive messages)
	go client.WritePump()

	// Register client with hub
	c.hub.Register <- client

	// Send connected event
	onlineUsers := c.hub.GetOnlineUsersInCourse(courseID)
	connectedEvent := ws.NewConnectedEvent(userID, courseID, onlineUsers)
	client.SendEvent(connectedEvent)

	// Start Redis subscription for this course (if not already subscribed)
	c.ensureRedisSubscription(courseID)

	// Start read pump (this blocks until connection closes)
	client.ReadPump(c.handleMessage)
}

// ensureRedisSubscription ensures we have exactly one Redis subscription per course
func (c *ChatWebSocketController) ensureRedisSubscription(courseID uint64) {
	c.subscribeMu.Lock()
	defer c.subscribeMu.Unlock()

	if c.subscribedCourses[courseID] {
		return // Already subscribed
	}

	c.subscribedCourses[courseID] = true
	go c.subscribeToRedisChannel(courseID)
}

// handleMessage processes incoming WebSocket messages from clients
func (c *ChatWebSocketController) handleMessage(client *ws.Client, envelope *ws.WSEnvelope) error {
	switch envelope.Type {
	case ws.EventTypeMessage:
		return c.handleSendMessage(client, envelope)
	case ws.EventTypeTyping:
		return c.handleTyping(client, envelope)
	case ws.EventTypeRead:
		return c.handleReadReceipt(client, envelope)
	default:
		return nil
	}
}

// handleSendMessage processes message send requests
func (c *ChatWebSocketController) handleSendMessage(client *ws.Client, envelope *ws.WSEnvelope) error {
	// Parse payload
	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	var payload ws.MessagePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return err
	}

	// Validate message
	if payload.Content == "" && len(payload.MediaIDs) == 0 {
		return ws.ErrEmptyMessage
	}

	if len(payload.Content) > 2000 {
		return ws.ErrMessageTooLong
	}

	// Create a mock gin.Context for service call
	fakeCtx := createFakeGinContext()

	var response *dto.ChatMessageResponse

	// Check if this is a reply or a new message
	if payload.ReplyToMessageID != nil && *payload.ReplyToMessageID > 0 {
		// Send as reply using SendReplyWithMedias which returns ChatMessageResponse
		replyReq := dto.SendChatMessageRequest{
			Content:  payload.Content,
			MediaIDs: payload.MediaIDs,
		}
		response, err = c.replyService.SendReplyWithMedias(
			fakeCtx.Request.Context(),
			client.CourseID,
			client.UserID,
			*payload.ReplyToMessageID,
			replyReq,
		)
		if err != nil {
			return err
		}

		// Broadcast reply event to all clients except sender
		replyEvent := ws.NewReplyEvent(response, *payload.ReplyToMessageID)
		c.broadcastToCourse(client.CourseID, replyEvent, client)

	} else {
		// Send as new message
		msgReq := dto.SendChatMessageRequest{
			Content:  payload.Content,
			MediaIDs: payload.MediaIDs,
		}

		if len(payload.MediaIDs) > 0 {
			response, err = c.chatService.SendMessageWithMedias(fakeCtx, client.CourseID, client.UserID, msgReq)
		} else {
			response, err = c.chatService.SendMessage(fakeCtx, client.CourseID, client.UserID, msgReq)
		}

		if err != nil {
			return err
		}

		// Note: The chat service already publishes to Redis, which will broadcast to other WS clients
		// So we don't need to manually broadcast here
	}

	// Send confirmation to sender (with message ID)
	confirmEvent := ws.NewMessageEvent(response)
	return client.SendEvent(confirmEvent)
}

// handleTyping processes typing indicator
func (c *ChatWebSocketController) handleTyping(client *ws.Client, envelope *ws.WSEnvelope) error {
	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	var payload ws.TypingPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return err
	}

	// Broadcast typing event to all clients except sender
	typingEvent := ws.NewTypingEvent(client.UserID, client.CourseID, payload.IsTyping)
	c.broadcastToCourse(client.CourseID, typingEvent, client)

	return nil
}

// handleReadReceipt processes read receipts
func (c *ChatWebSocketController) handleReadReceipt(client *ws.Client, envelope *ws.WSEnvelope) error {
	// Read receipts can be stored in DB and/or broadcast
	// For now, just return nil - can be extended later
	return nil
}

// broadcastToCourse broadcasts an event to all clients in a course
func (c *ChatWebSocketController) broadcastToCourse(courseID uint64, event ws.WSEnvelope, excludeClient *ws.Client) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	c.hub.Broadcast <- &ws.BroadcastMessage{
		CourseID: courseID,
		Data:     data,
	}
}

// subscribeToRedisChannel subscribes to Redis pub/sub for a course
// This allows messages from REST API to be broadcasted to WebSocket clients
func (c *ChatWebSocketController) subscribeToRedisChannel(courseID uint64) {
	messageChan, cleanup, err := c.pubsubService.SubscribeToCourse(courseID)
	if err != nil {
		// Remove from subscribed courses so we can retry later
		c.subscribeMu.Lock()
		delete(c.subscribedCourses, courseID)
		c.subscribeMu.Unlock()
		return
	}
	defer cleanup()

	for msg := range messageChan {
		if msg == nil {
			continue
		}

		// Convert Redis message to WebSocket event
		event := c.convertRedisMessageToWSEvent(msg)
		if event == nil {
			continue
		}

		// Broadcast to WebSocket clients
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		c.hub.BroadcastToCourseFromRedis(courseID, data)
	}

	// Channel closed, remove from subscribed courses
	c.subscribeMu.Lock()
	delete(c.subscribedCourses, courseID)
	c.subscribeMu.Unlock()
}

// convertRedisMessageToWSEvent converts a Redis pub/sub message to a WebSocket event
func (c *ChatWebSocketController) convertRedisMessageToWSEvent(msg *redis.ChatPubSubMessage) *ws.WSEnvelope {
	switch msg.Type {
	case "message":
		return &ws.WSEnvelope{
			Type:    ws.EventTypeMessage,
			Payload: msg.Data,
		}

	case "reply":
		return &ws.WSEnvelope{
			Type:    ws.EventTypeReply,
			Payload: msg.Data,
		}

	case "pin":
		return &ws.WSEnvelope{
			Type:    ws.EventTypePin,
			Payload: msg.Data,
		}

	case "delete":
		return &ws.WSEnvelope{
			Type:    ws.EventTypeDelete,
			Payload: msg.Data,
		}

	case "private_message":
		return &ws.WSEnvelope{
			Type:    ws.EventTypePrivate,
			Payload: msg.Data,
		}

	default:
		return nil
	}
}

// createFakeGinContext creates a minimal gin.Context for service calls
func createFakeGinContext() *gin.Context {
	ctx := &gin.Context{}
	ctx.Request, _ = http.NewRequest("POST", "/", nil)
	return ctx
}

// GetOnlineUsers returns the list of online users in a course
func (c *ChatWebSocketController) GetOnlineUsers(ctx *gin.Context) {
	courseIDStr := ctx.Param("id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	onlineUsers := c.hub.GetOnlineUsersInCourse(courseID)

	ctx.JSON(http.StatusOK, gin.H{
		"course_id":    courseID,
		"online_users": onlineUsers,
		"count":        len(onlineUsers),
	})
}

// GetConnectionStats returns WebSocket connection statistics
func (c *ChatWebSocketController) GetConnectionStats(ctx *gin.Context) {
	courseIDStr := ctx.Query("course_id")

	if courseIDStr != "" {
		courseID, err := strconv.ParseUint(courseIDStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"course_id":    courseID,
			"client_count": c.hub.GetCourseClientCount(courseID),
			"online_users": c.hub.GetOnlineUsersInCourse(courseID),
		})
		return
	}

	// Return general stats
	ctx.JSON(http.StatusOK, gin.H{
		"status": "WebSocket hub running",
	})
}
