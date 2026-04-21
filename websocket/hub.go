package websocket

import (
	"be-lms/config"
	"be-lms/redis"
	"encoding/json"
	"sync"
	"time"
)

// Client represents a single websocket connection
type Client struct {
	Hub       *Hub
	Conn      *SafeConn
	Send      chan []byte
	done      chan struct{}
	closeOnce sync.Once
	UserID    uint64
	CourseID  uint64
	JoinedAt  time.Time
}

// NewClient creates a new Client with initialized channels
func NewClient(hub *Hub, conn *SafeConn, userID, courseID uint64) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		done:     make(chan struct{}),
		UserID:   userID,
		CourseID: courseID,
		JoinedAt: time.Now(),
	}
}

// Close safely closes the client, preventing further sends
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		// Check if done channel is initialized before closing
		if c.done != nil {
			close(c.done)
		}
		// Don't close c.Send here - let WritePump close it when it exits
		// This prevents panic if Hub is still trying to send
	})
}

// TrySend attempts to send a message without blocking or panicking
// Returns false if client is closed or buffer is full
func (c *Client) TrySend(msg []byte) bool {
	// Check if client is closed
	select {
	case <-c.done:
		return false
	default:
	}

	// Try to send (non-blocking)
	select {
	case c.Send <- msg:
		return true
	default:
		// Buffer full
		return false
	}
}

// ClientLifecycleCallback is called when a client joins or leaves a course
type ClientLifecycleCallback func(courseID uint64, isJoin bool)

// Hub maintains the set of active clients and broadcasts messages
// All state access must go through the Run() goroutine (actor pattern)
type Hub struct {
	// Registered clients by course (only accessed in Run() goroutine)
	courseClients map[uint64]map[*Client]bool

	// All clients by user (only accessed in Run() goroutine)
	userClients map[uint64]map[*Client]bool

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Inbound messages from clients (all broadcasts go through this)
	Broadcast chan *BroadcastMessage

	// Chat pubsub service for Redis integration
	pubsub *redis.ChatPubSub

	// Lifecycle callbacks for course subscription management
	lifecycleCallbacks []ClientLifecycleCallback
	callbackMu         sync.RWMutex

	// Mutex for read-only operations (GetCourseClientCount, GetOnlineUsersInCourse)
	// These are called from outside Run() goroutine, so need protection
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	CourseID      uint64
	UserID        uint64 // Optional: for user-specific broadcasts
	Type          string
	Data          []byte
	ExcludeUserID uint64 // Optional: exclude this user from broadcast (for join/leave events)
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		courseClients:      make(map[uint64]map[*Client]bool),
		userClients:        make(map[uint64]map[*Client]bool),
		Register:           make(chan *Client),
		Unregister:         make(chan *Client),
		Broadcast:          make(chan *BroadcastMessage, 256),
		pubsub:             redis.GetChatPubSub(),
		lifecycleCallbacks: make([]ClientLifecycleCallback, 0),
	}
}

// OnClientLifecycle registers a callback for client join/leave events
func (h *Hub) OnClientLifecycle(callback ClientLifecycleCallback) {
	h.callbackMu.Lock()
	defer h.callbackMu.Unlock()
	h.lifecycleCallbacks = append(h.lifecycleCallbacks, callback)
}

// notifyLifecycle notifies all registered callbacks
func (h *Hub) notifyLifecycle(courseID uint64, isJoin bool) {
	h.callbackMu.RLock()
	callbacks := make([]ClientLifecycleCallback, len(h.lifecycleCallbacks))
	copy(callbacks, h.lifecycleCallbacks)
	h.callbackMu.RUnlock()

	for _, cb := range callbacks {
		go cb(courseID, isJoin) // Run in goroutine to avoid blocking Hub
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub
// Must only be called from Run() goroutine (no locking needed)
func (h *Hub) registerClient(client *Client) {
	// Check if this is the first client in the course
	isFirstClient := h.courseClients[client.CourseID] == nil || len(h.courseClients[client.CourseID]) == 0

	// Add to course clients
	if h.courseClients[client.CourseID] == nil {
		h.courseClients[client.CourseID] = make(map[*Client]bool)
	}
	h.courseClients[client.CourseID][client] = true

	// Add to user clients
	if h.userClients[client.UserID] == nil {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true

	config.Log.Infof("👤 [Hub] Client registered: user_id=%d, course_id=%d, total_in_course=%d, is_first=%v", client.UserID, client.CourseID, len(h.courseClients[client.CourseID]), isFirstClient)

	// Send join notification to other clients via channel (actor pattern)
	joinEvent := WSEnvelope{
		Type: EventTypeJoin,
		Payload: map[string]interface{}{
			"user_id":   client.UserID,
			"course_id": client.CourseID,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(joinEvent)
	if err == nil {
		// Send via channel to ensure it's processed in Run() goroutine
		// Exclude the joining user from receiving their own join notification
		select {
		case h.Broadcast <- &BroadcastMessage{
			CourseID:      client.CourseID,
			Data:          data,
			ExcludeUserID: client.UserID,
		}:
		default:
			// Channel full, skip (non-critical event)
		}
	}

	// Notify lifecycle callbacks if this is the first client
	if isFirstClient {
		h.notifyLifecycle(client.CourseID, true)
	}
}

// unregisterClient removes a client from the hub
// Must only be called from Run() goroutine (no locking needed)
func (h *Hub) unregisterClient(client *Client) {
	isLastClient := false
	courseID := client.CourseID

	// Remove from course clients
	if clients, ok := h.courseClients[client.CourseID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.courseClients, client.CourseID)
				isLastClient = true
			}
		}
	}

	// Remove from user clients
	if clients, ok := h.userClients[client.UserID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}
	}

	// Mark client as closed (prevents further sends)
	client.Close()

	config.Log.Infof("👋 [Hub] Client unregistered: user_id=%d, course_id=%d, remaining_in_course=%d, is_last=%v", client.UserID, courseID, len(h.courseClients[courseID]), isLastClient)

	// Send leave notification to other clients via channel (actor pattern)
	leaveEvent := WSEnvelope{
		Type: EventTypeLeave,
		Payload: map[string]interface{}{
			"user_id":   client.UserID,
			"course_id": client.CourseID,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(leaveEvent)
	if err == nil {
		// Send via channel to ensure it's processed in Run() goroutine
		select {
		case h.Broadcast <- &BroadcastMessage{
			CourseID:      client.CourseID,
			Data:          data,
			ExcludeUserID: client.UserID, // Exclude leaving user (though they're already removed)
		}:
		default:
			// Channel full, skip (non-critical event)
		}
	}

	// Notify lifecycle callbacks if this was the last client
	if isLastClient {
		h.notifyLifecycle(courseID, false)
	}
}

// broadcastMessage sends a message to all clients in a course or user
// Must only be called from Run() goroutine (no locking needed)
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	sentCount := 0
	failedCount := 0

	if message.Type == "user_broadcast" {
		// Broadcast to all connections of a specific user
		if clients, ok := h.userClients[message.UserID]; ok {
			for client := range clients {
				if !client.TrySend(message.Data) {
					failedCount++
					// Client buffer full or closed, queue for unregister
					select {
					case h.Unregister <- client:
					default:
						// Unregister channel full, skip
					}
				} else {
					sentCount++
				}
			}
		}
		config.Log.Infof("📡 [Hub] User broadcast: user_id=%d, sent=%d, failed=%d", message.UserID, sentCount, failedCount)
	} else {
		// Broadcast to all clients in a course
		if clients, ok := h.courseClients[message.CourseID]; ok {
			for client := range clients {
				// Exclude specific user if specified
				if message.ExcludeUserID > 0 && client.UserID == message.ExcludeUserID {
					continue
				}

				if !client.TrySend(message.Data) {
					failedCount++
					// Client buffer full or closed, queue for unregister
					select {
					case h.Unregister <- client:
					default:
						// Unregister channel full, skip
					}
				} else {
					sentCount++
				}
			}
		}
		config.Log.Infof("📡 [Hub] Course broadcast: course_id=%d, exclude_user_id=%d, sent=%d, failed=%d", message.CourseID, message.ExcludeUserID, sentCount, failedCount)
	}
}

// BroadcastToCourse sends an event to all clients in a course (except sender)
// Thread-safe: sends via channel to Hub goroutine
func (h *Hub) BroadcastToCourse(courseID uint64, event WSEnvelope, excludeUserID uint64) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	select {
	case h.Broadcast <- &BroadcastMessage{
		CourseID:      courseID,
		Data:          data,
		ExcludeUserID: excludeUserID,
	}:
	default:
		// Channel full, skip (non-blocking)
	}
}

// BroadcastToCourseFromRedis broadcasts a message from Redis to all WS clients
// Thread-safe: sends via channel to Hub goroutine
func (h *Hub) BroadcastToCourseFromRedis(courseID uint64, data []byte) {
	select {
	case h.Broadcast <- &BroadcastMessage{
		CourseID: courseID,
		Data:     data,
	}:
	default:
		// Channel full, skip (non-blocking)
	}
}

// BroadcastToUser sends an event to all connections of a specific user
// Thread-safe: sends via channel to Hub goroutine
func (h *Hub) BroadcastToUser(userID uint64, event WSEnvelope) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// For user-specific broadcasts, we need to check userClients
	// Since we can't do this directly, we'll need to add a new message type
	// For now, broadcast to all courses the user is in
	// This is a limitation but ensures thread-safety
	select {
	case h.Broadcast <- &BroadcastMessage{
		UserID: userID,
		Data:   data,
		Type:   "user_broadcast", // Special type for user broadcasts
	}:
	default:
		// Channel full, skip (non-blocking)
	}
}

// GetCourseClientCount returns the number of clients in a course
func (h *Hub) GetCourseClientCount(courseID uint64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.courseClients[courseID]; ok {
		return len(clients)
	}
	return 0
}

// GetUserClientCount returns the number of connections for a user
func (h *Hub) GetUserClientCount(userID uint64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.userClients[userID]; ok {
		return len(clients)
	}
	return 0
}

// GetOnlineUsersInCourse returns list of online user IDs in a course
func (h *Hub) GetOnlineUsersInCourse(courseID uint64) []uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userSet := make(map[uint64]bool)
	if clients, ok := h.courseClients[courseID]; ok {
		for client := range clients {
			userSet[client.UserID] = true
		}
	}

	users := make([]uint64, 0, len(userSet))
	for userID := range userSet {
		users = append(users, userID)
	}
	return users
}

// Global hub instance
var globalHub *Hub
var hubOnce sync.Once

// GetHub returns the global hub instance
func GetHub() *Hub {
	hubOnce.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
		config.Log.Info("🚀 WebSocket Hub started")
	})
	return globalHub
}
