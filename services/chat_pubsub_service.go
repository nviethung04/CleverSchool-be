package services

import (
	"be-Clever School/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ChatPubSubService handles real-time chat notifications via Redis pub/sub
type ChatPubSubService interface {
	// Message events
	NotifyNewMessage(courseID, messageID, userID uint64, messageData interface{}) error
	NotifyNewReply(courseID, messageID, parentMessageID, userID uint64, replyData interface{}) error
	NotifyMessageDeleted(courseID, messageID, userID uint64) error
	NotifyMessagePinned(courseID, messageID, userID uint64, isPinned bool) error
	NotifyNewPrivateMessage(courseID, messageID, senderID, recipientID uint64, messageData interface{}) error

	// Reaction events
	NotifyReactionAdded(courseID, messageID, userID uint64, reactionID uint64, emoji string, totalCount int) error
	NotifyReactionRemoved(courseID, messageID, userID uint64, emoji string, totalCount int) error
	NotifyAllReactionsRemoved(courseID, messageID, userID uint64) error

	// User notifications
	NotifyUser(userID uint64, notificationType string, data interface{}) error

	// Subscription management
	SubscribeToCourse(courseID uint64) (<-chan *redis.ChatPubSubMessage, func(), error)
	SubscribeToUser(userID uint64) (<-chan *redis.ChatPubSubMessage, func(), error)
	SubscribeToMultipleCourses(courseIDs []uint64) (<-chan *redis.ChatPubSubMessage, func(), error)
	SubscribeToPrivateMessages(courseID, userID1, userID2 uint64) (<-chan *redis.ChatPubSubMessage, func(), error)
	SubscribeToAllPrivateMessages(courseID, userID uint64) (<-chan *redis.ChatPubSubMessage, func(), error)

	// Monitoring
	GetActiveChannels() ([]string, error)
	GetChannelSubscribers(channel string) (int64, error)
}

type chatPubSubService struct {
	pubsub *redis.ChatPubSub
}

// NewChatPubSubService creates a new ChatPubSubService
func NewChatPubSubService() ChatPubSubService {
	return &chatPubSubService{
		pubsub: redis.GetChatPubSub(),
	}
}

// NotifyNewMessage publishes a new message event
func (s *chatPubSubService) NotifyNewMessage(courseID, messageID, userID uint64, messageData interface{}) error {
	log.Printf("Publishing new message event: courseID=%d, messageID=%d, userID=%d", courseID, messageID, userID)
	return s.pubsub.PublishNewMessage(courseID, messageID, userID, messageData)
}

// NotifyNewReply publishes a new reply event
func (s *chatPubSubService) NotifyNewReply(courseID, messageID, parentMessageID, userID uint64, replyData interface{}) error {
	log.Printf("Publishing new reply event: courseID=%d, messageID=%d, parentID=%d, userID=%d", courseID, messageID, parentMessageID, userID)
	return s.pubsub.PublishNewReply(courseID, messageID, parentMessageID, userID, replyData)
}

// NotifyMessageDeleted publishes a message deletion event
func (s *chatPubSubService) NotifyMessageDeleted(courseID, messageID, userID uint64) error {
	log.Printf("Publishing message deleted event: courseID=%d, messageID=%d, userID=%d", courseID, messageID, userID)
	return s.pubsub.PublishDelete(courseID, messageID, userID)
}

// NotifyMessagePinned publishes a message pin/unpin event
func (s *chatPubSubService) NotifyMessagePinned(courseID, messageID, userID uint64, isPinned bool) error {
	log.Printf("Publishing message pin event: courseID=%d, messageID=%d, userID=%d, pinned=%t", courseID, messageID, userID, isPinned)
	return s.pubsub.PublishPin(courseID, messageID, userID, isPinned)
}

// NotifyReactionAdded publishes a reaction added event
func (s *chatPubSubService) NotifyReactionAdded(courseID, messageID, userID uint64, reactionID uint64, emoji string, totalCount int) error {
	reaction := redis.ChatReactionEvent{
		Action:        "add",
		ReactionID:    reactionID,
		MessageID:     messageID,
		UserID:        userID,
		Emoji:         emoji,
		ReactionCount: totalCount,
	}

	log.Printf("Publishing reaction added event: courseID=%d, messageID=%d, userID=%d, emoji=%s", courseID, messageID, userID, emoji)
	return s.pubsub.PublishReaction(courseID, messageID, userID, reaction)
}

// NotifyReactionRemoved publishes a reaction removed event
func (s *chatPubSubService) NotifyReactionRemoved(courseID, messageID, userID uint64, emoji string, totalCount int) error {
	reaction := redis.ChatReactionEvent{
		Action:        "remove",
		MessageID:     messageID,
		UserID:        userID,
		Emoji:         emoji,
		ReactionCount: totalCount,
	}

	log.Printf("Publishing reaction removed event: courseID=%d, messageID=%d, userID=%d, emoji=%s", courseID, messageID, userID, emoji)
	return s.pubsub.PublishReaction(courseID, messageID, userID, reaction)
}

// NotifyAllReactionsRemoved publishes an event when all reactions are removed
func (s *chatPubSubService) NotifyAllReactionsRemoved(courseID, messageID, userID uint64) error {
	reaction := redis.ChatReactionEvent{
		Action:        "remove_all",
		MessageID:     messageID,
		UserID:        userID,
		ReactionCount: 0,
	}

	log.Printf("Publishing all reactions removed event: courseID=%d, messageID=%d, userID=%d", courseID, messageID, userID)
	return s.pubsub.PublishReaction(courseID, messageID, userID, reaction)
}

// NotifyUser sends a notification to a specific user
func (s *chatPubSubService) NotifyUser(userID uint64, notificationType string, data interface{}) error {
	notification := map[string]interface{}{
		"type": notificationType,
		"data": data,
	}

	log.Printf("Publishing user notification: userID=%d, type=%s", userID, notificationType)
	return s.pubsub.PublishUserNotification(userID, notification)
}

// NotifyNewPrivateMessage publishes a new private message event
func (s *chatPubSubService) NotifyNewPrivateMessage(courseID, messageID, senderID, recipientID uint64, messageData interface{}) error {
	log.Printf("Publishing new private message event: courseID=%d, messageID=%d, senderID=%d, recipientID=%d", courseID, messageID, senderID, recipientID)
	return s.pubsub.PublishPrivateMessage(courseID, messageID, senderID, recipientID, messageData)
}

// SubscribeToCourse subscribes to course chat events
func (s *chatPubSubService) SubscribeToCourse(courseID uint64) (<-chan *redis.ChatPubSubMessage, func(), error) {
	log.Printf("Subscribing to course chat: courseID=%d", courseID)
	return s.pubsub.SubscribeToCourse(courseID)
}

// SubscribeToUser subscribes to user notifications
func (s *chatPubSubService) SubscribeToUser(userID uint64) (<-chan *redis.ChatPubSubMessage, func(), error) {
	log.Printf("Subscribing to user notifications: userID=%d", userID)
	return s.pubsub.SubscribeToUser(userID)
}

// SubscribeToMultipleCourses subscribes to multiple course chat events
func (s *chatPubSubService) SubscribeToMultipleCourses(courseIDs []uint64) (<-chan *redis.ChatPubSubMessage, func(), error) {
	log.Printf("Subscribing to multiple courses: %v", courseIDs)
	return s.pubsub.SubscribeToMultipleCourses(courseIDs)
}

// SubscribeToPrivateMessages subscribes to private messages between two users in a course
func (s *chatPubSubService) SubscribeToPrivateMessages(courseID, userID1, userID2 uint64) (<-chan *redis.ChatPubSubMessage, func(), error) {
	log.Printf("Subscribing to private messages: courseID=%d, userID1=%d, userID2=%d", courseID, userID1, userID2)
	return s.pubsub.SubscribeToPrivateMessages(courseID, userID1, userID2)
}

// SubscribeToAllPrivateMessages subscribes to all private messages for a user in a course
func (s *chatPubSubService) SubscribeToAllPrivateMessages(courseID, userID uint64) (<-chan *redis.ChatPubSubMessage, func(), error) {
	log.Printf("Subscribing to all private messages: courseID=%d, userID=%d", courseID, userID)
	return s.pubsub.SubscribeToAllPrivateMessages(courseID, userID)
}

// GetActiveChannels returns active chat channels
func (s *chatPubSubService) GetActiveChannels() ([]string, error) {
	return s.pubsub.GetActiveChannels()
}

// GetChannelSubscribers returns subscriber count for a channel
func (s *chatPubSubService) GetChannelSubscribers(channel string) (int64, error) {
	return s.pubsub.GetChannelSubscribers(channel)
}

// ChatEventHandler handles incoming chat events from pub/sub
type ChatEventHandler struct {
	onNewMessage   func(courseID, messageID, userID uint64, data interface{})
	onNewReply     func(courseID, messageID, parentMessageID, userID uint64, data interface{})
	onReaction     func(courseID, messageID, userID uint64, reaction redis.ChatReactionEvent)
	onPin          func(courseID, messageID, userID uint64, isPinned bool)
	onDelete       func(courseID, messageID, userID uint64)
	onNotification func(userID uint64, notification interface{})
}

// NewChatEventHandler creates a new event handler
func NewChatEventHandler() *ChatEventHandler {
	return &ChatEventHandler{}
}

// OnNewMessage sets the handler for new message events
func (h *ChatEventHandler) OnNewMessage(handler func(courseID, messageID, userID uint64, data interface{})) *ChatEventHandler {
	h.onNewMessage = handler
	return h
}

// OnNewReply sets the handler for new reply events
func (h *ChatEventHandler) OnNewReply(handler func(courseID, messageID, parentMessageID, userID uint64, data interface{})) *ChatEventHandler {
	h.onNewReply = handler
	return h
}

// OnReaction sets the handler for reaction events
func (h *ChatEventHandler) OnReaction(handler func(courseID, messageID, userID uint64, reaction redis.ChatReactionEvent)) *ChatEventHandler {
	h.onReaction = handler
	return h
}

// OnPin sets the handler for pin/unpin events
func (h *ChatEventHandler) OnPin(handler func(courseID, messageID, userID uint64, isPinned bool)) *ChatEventHandler {
	h.onPin = handler
	return h
}

// OnDelete sets the handler for delete events
func (h *ChatEventHandler) OnDelete(handler func(courseID, messageID, userID uint64)) *ChatEventHandler {
	h.onDelete = handler
	return h
}

// OnNotification sets the handler for user notifications
func (h *ChatEventHandler) OnNotification(handler func(userID uint64, notification interface{})) *ChatEventHandler {
	h.onNotification = handler
	return h
}

// HandleMessage processes incoming pub/sub messages
func (h *ChatEventHandler) HandleMessage(msg *redis.ChatPubSubMessage) {
	switch msg.Type {
	case "message":
		if h.onNewMessage != nil {
			h.onNewMessage(msg.CourseID, msg.MessageID, msg.UserID, msg.Data)
		}

	case "reply":
		if h.onNewReply != nil && msg.Data != nil {
			if dataMap, ok := msg.Data.(map[string]interface{}); ok {
				if parentIDFloat, ok := dataMap["parent_message_id"].(float64); ok {
					parentID := uint64(parentIDFloat)
					h.onNewReply(msg.CourseID, msg.MessageID, parentID, msg.UserID, dataMap["reply"])
				}
			}
		}

	case "reaction":
		if h.onReaction != nil && msg.Data != nil {
			// Convert interface{} back to ChatReactionEvent
			if dataBytes, err := json.Marshal(msg.Data); err == nil {
				var reaction redis.ChatReactionEvent
				if err := json.Unmarshal(dataBytes, &reaction); err == nil {
					h.onReaction(msg.CourseID, msg.MessageID, msg.UserID, reaction)
				}
			}
		}

	case "pin":
		if h.onPin != nil && msg.Data != nil {
			if dataMap, ok := msg.Data.(map[string]interface{}); ok {
				if isPinnedBool, ok := dataMap["is_pinned"].(bool); ok {
					h.onPin(msg.CourseID, msg.MessageID, msg.UserID, isPinnedBool)
				}
			}
		}

	case "delete":
		if h.onDelete != nil {
			h.onDelete(msg.CourseID, msg.MessageID, msg.UserID)
		}

	case "notification":
		if h.onNotification != nil {
			h.onNotification(msg.UserID, msg.Data)
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// StartChatEventListener starts listening for chat events and processes them with the handler
func StartChatEventListener(ctx context.Context, pubsubService ChatPubSubService, courseIDs []uint64, handler *ChatEventHandler) error {
	if len(courseIDs) == 0 {
		return fmt.Errorf("no course IDs provided")
	}

	messageChan, cleanup, err := pubsubService.SubscribeToMultipleCourses(courseIDs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to courses: %w", err)
	}

	go func() {
		defer cleanup()
		for {
			select {
			case msg, ok := <-messageChan:
				if !ok {
					log.Println("Chat event channel closed")
					return
				}
				if msg != nil {
					handler.HandleMessage(msg)
				}
			case <-ctx.Done():
				log.Println("Chat event listener stopped")
				return
			}
		}
	}()

	log.Printf("Started chat event listener for %d courses", len(courseIDs))
	return nil
}
