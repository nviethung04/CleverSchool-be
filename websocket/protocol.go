package websocket

import (
	"encoding/json"
	"time"
)

// Event types - SHARED protocol between BE and FE
const (
	// Client -> Server events
	EventTypeMessage = "message"
	EventTypeReply   = "reply"
	EventTypeTyping  = "typing"
	EventTypeRead    = "read"

	// Server -> Client events (also used for broadcasts)
	EventTypeJoin         = "join"
	EventTypeLeave        = "leave"
	EventTypePin          = "pin"
	EventTypeDelete       = "delete"
	EventTypePrivate      = "private_message"
	EventTypeConnected    = "connected"
	EventTypeHeartbeat    = "heartbeat"
	EventTypeError        = "error"
	EventTypeNotification = "notification"
)

// WSEnvelope is the standard message envelope for all WebSocket communications
// This is the SHARED protocol between Backend and Frontend
type WSEnvelope struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ===== CLIENT -> SERVER PAYLOADS =====

// MessagePayload is sent when client wants to send a new message
type MessagePayload struct {
	Content          string  `json:"content"`
	ReplyToMessageID *uint64 `json:"reply_to_message_id,omitempty"`
	MediaIDs         []int64 `json:"media_ids,omitempty"`
}

// TypingPayload is sent when client starts/stops typing
type TypingPayload struct {
	IsTyping bool `json:"is_typing"`
}

// ReadPayload is sent when client marks messages as read
type ReadPayload struct {
	MessageIDs []uint64 `json:"message_ids"`
}

// ===== SERVER -> CLIENT PAYLOADS =====

// ConnectedPayload is sent when connection is established
type ConnectedPayload struct {
	UserID      uint64   `json:"user_id"`
	CourseID    uint64   `json:"course_id"`
	OnlineUsers []uint64 `json:"online_users,omitempty"`
	Timestamp   string   `json:"timestamp"`
}

// HeartbeatPayload is sent periodically to keep connection alive
type HeartbeatPayload struct {
	Timestamp string `json:"timestamp"`
}

// ErrorPayload is sent when an error occurs
type ErrorPayload struct {
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	Timestamp string `json:"timestamp"`
}

// JoinPayload is sent when a user joins the chat
type JoinPayload struct {
	UserID    uint64 `json:"user_id"`
	CourseID  uint64 `json:"course_id"`
	Timestamp string `json:"timestamp"`
}

// LeavePayload is sent when a user leaves the chat
type LeavePayload struct {
	UserID    uint64 `json:"user_id"`
	CourseID  uint64 `json:"course_id"`
	Timestamp string `json:"timestamp"`
}

// TypingEventPayload is broadcast when user is typing
type TypingEventPayload struct {
	UserID    uint64 `json:"user_id"`
	CourseID  uint64 `json:"course_id"`
	IsTyping  bool   `json:"is_typing"`
	Timestamp string `json:"timestamp"`
}

// ===== HELPER FUNCTIONS =====

// ParseEnvelope parses incoming WebSocket message into an envelope
func ParseEnvelope(data []byte) (*WSEnvelope, error) {
	var envelope WSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// NewConnectedEvent creates a connected event envelope
func NewConnectedEvent(userID, courseID uint64, onlineUsers []uint64) WSEnvelope {
	return WSEnvelope{
		Type: EventTypeConnected,
		Payload: ConnectedPayload{
			UserID:      userID,
			CourseID:    courseID,
			OnlineUsers: onlineUsers,
			Timestamp:   time.Now().Format(time.RFC3339),
		},
	}
}

// NewHeartbeatEvent creates a heartbeat event envelope
func NewHeartbeatEvent() WSEnvelope {
	return WSEnvelope{
		Type: EventTypeHeartbeat,
		Payload: HeartbeatPayload{
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

// NewErrorEvent creates an error event envelope
func NewErrorEvent(message string, code string) WSEnvelope {
	return WSEnvelope{
		Type: EventTypeError,
		Payload: ErrorPayload{
			Message:   message,
			Code:      code,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

// NewMessageEvent creates a message event with ChatMessageResponse payload
func NewMessageEvent(payload interface{}) WSEnvelope {
	return WSEnvelope{
		Type:    EventTypeMessage,
		Payload: payload,
	}
}

// NewReplyEvent creates a reply event with reply data and parent ID
func NewReplyEvent(reply interface{}, parentMessageID uint64) WSEnvelope {
	return WSEnvelope{
		Type: EventTypeReply,
		Payload: map[string]interface{}{
			"reply":             reply,
			"parent_message_id": parentMessageID,
		},
	}
}

// NewPinEvent creates a pin/unpin event
func NewPinEvent(messageID, userID uint64, isPinned bool) WSEnvelope {
	return WSEnvelope{
		Type: EventTypePin,
		Payload: map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
			"is_pinned":  isPinned,
		},
	}
}

// NewDeleteEvent creates a delete message event
func NewDeleteEvent(messageID, userID uint64) WSEnvelope {
	return WSEnvelope{
		Type: EventTypeDelete,
		Payload: map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
		},
	}
}

// NewTypingEvent creates a typing event
func NewTypingEvent(userID, courseID uint64, isTyping bool) WSEnvelope {
	return WSEnvelope{
		Type: EventTypeTyping,
		Payload: TypingEventPayload{
			UserID:    userID,
			CourseID:  courseID,
			IsTyping:  isTyping,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

// NewPrivateMessageEvent creates a private message event
func NewPrivateMessageEvent(message interface{}, recipientID uint64) WSEnvelope {
	return WSEnvelope{
		Type: EventTypePrivate,
		Payload: map[string]interface{}{
			"message":      message,
			"recipient_id": recipientID,
		},
	}
}

