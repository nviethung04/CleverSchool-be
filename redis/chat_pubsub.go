package redis

import (
	"be-cleverschool/database/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ChatMessage represents a chat message for pub/sub
type ChatPubSubMessage struct {
	Type      string      `json:"type"` // "message", "reply", "reaction", "pin", "delete"
	CourseID  uint64      `json:"course_id"`
	MessageID uint64      `json:"message_id"`
	UserID    uint64      `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ChatReactionEvent represents a reaction event
type ChatReactionEvent struct {
	Action        string `json:"action"` // "add", "remove", "remove_all"
	ReactionID    uint64 `json:"reaction_id"`
	MessageID     uint64 `json:"message_id"`
	UserID        uint64 `json:"user_id"`
	Emoji         string `json:"emoji"`
	ReactionCount int    `json:"reaction_count"`
}

// ChatPinEvent represents a pin/unpin event
type ChatPinEvent struct {
	MessageID uint64 `json:"message_id"`
	IsPinned  bool   `json:"is_pinned"`
	UserID    uint64 `json:"user_id"`
}

// ChatDeleteEvent represents a delete event
type ChatDeleteEvent struct {
	MessageID uint64 `json:"message_id"`
	UserID    uint64 `json:"user_id"`
}

// ChatPubSub manages Redis pub/sub for chat functionality
type ChatPubSub struct {
	client *redis.Client
	ctx    context.Context
}

// NewChatPubSub creates a new ChatPubSub instance
func NewChatPubSub() *ChatPubSub {
	return &ChatPubSub{
		client: db.RedisClient,
		ctx:    context.Background(),
	}
}

// GetCourseChannel returns the Redis channel name for a course
func (c *ChatPubSub) GetCourseChannel(courseID uint64) string {
	return fmt.Sprintf("chat:course:%d", courseID)
}

// GetUserChannel returns the Redis channel name for a user (for notifications)
func (c *ChatPubSub) GetUserChannel(userID uint64) string {
	return fmt.Sprintf("chat:user:%d", userID)
}

// GetPrivateMessageChannel returns the Redis channel name for private messages between two users in a course
func (c *ChatPubSub) GetPrivateMessageChannel(courseID, userID1, userID2 uint64) string {
	// Sort user IDs to ensure consistent channel name regardless of sender/recipient order
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	return fmt.Sprintf("chat:private:course:%d:users:%d:%d", courseID, userID1, userID2)
}

// PublishNewMessage publishes a new message event
func (c *ChatPubSub) PublishNewMessage(courseID, messageID, userID uint64, messageData interface{}) error {
	message := ChatPubSubMessage{
		Type:      "message",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    userID,
		Data:      messageData,
		Timestamp: time.Now(),
	}

	return c.publishToCourse(courseID, message)
}

// PublishNewReply publishes a new reply event
func (c *ChatPubSub) PublishNewReply(courseID, messageID, parentMessageID, userID uint64, replyData interface{}) error {
	// Use distinct event type "reply" so frontend can differentiate
	message := ChatPubSubMessage{
		Type:      "reply",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    userID,
		Data: map[string]interface{}{
			"reply":             replyData,
			"parent_message_id": parentMessageID,
		},
		Timestamp: time.Now(),
	}

	return c.publishToCourse(courseID, message)
}

// PublishReaction publishes a reaction event
func (c *ChatPubSub) PublishReaction(courseID, messageID, userID uint64, reaction ChatReactionEvent) error {
	message := ChatPubSubMessage{
		Type:      "reaction",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    userID,
		Data:      reaction,
		Timestamp: time.Now(),
	}

	return c.publishToCourse(courseID, message)
}

// PublishPin publishes a pin/unpin event
func (c *ChatPubSub) PublishPin(courseID, messageID, userID uint64, isPinned bool) error {
	pinEvent := ChatPinEvent{
		MessageID: messageID,
		IsPinned:  isPinned,
		UserID:    userID,
	}

	message := ChatPubSubMessage{
		Type:      "pin",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    userID,
		Data:      pinEvent,
		Timestamp: time.Now(),
	}

	return c.publishToCourse(courseID, message)
}

// PublishDelete publishes a delete message event
func (c *ChatPubSub) PublishDelete(courseID, messageID, userID uint64) error {
	deleteEvent := ChatDeleteEvent{
		MessageID: messageID,
		UserID:    userID,
	}

	message := ChatPubSubMessage{
		Type:      "delete",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    userID,
		Data:      deleteEvent,
		Timestamp: time.Now(),
	}

	return c.publishToCourse(courseID, message)
}

// PublishUserNotification publishes a notification to a specific user
func (c *ChatPubSub) PublishUserNotification(userID uint64, notificationData interface{}) error {
	message := ChatPubSubMessage{
		Type:      "notification",
		UserID:    userID,
		Data:      notificationData,
		Timestamp: time.Now(),
	}

	return c.publishToUser(userID, message)
}

// PublishPrivateMessage publishes a private message event between two users in a course
func (c *ChatPubSub) PublishPrivateMessage(courseID, messageID, senderID, recipientID uint64, messageData interface{}) error {
	message := ChatPubSubMessage{
		Type:      "private_message",
		CourseID:  courseID,
		MessageID: messageID,
		UserID:    senderID,
		Data: map[string]interface{}{
			"message":      messageData,
			"recipient_id": recipientID,
		},
		Timestamp: time.Now(),
	}

	channel := c.GetPrivateMessageChannel(courseID, senderID, recipientID)
	return c.publish(channel, message)
}

// publishToCourse publishes a message to a course channel
func (c *ChatPubSub) publishToCourse(courseID uint64, message ChatPubSubMessage) error {
	channel := c.GetCourseChannel(courseID)
	return c.publish(channel, message)
}

// publishToUser publishes a message to a user channel
func (c *ChatPubSub) publishToUser(userID uint64, message ChatPubSubMessage) error {
	channel := c.GetUserChannel(userID)
	return c.publish(channel, message)
}

// publish publishes a message to a Redis channel
func (c *ChatPubSub) publish(channel string, message ChatPubSubMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.client.Publish(c.ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	log.Printf("Published message to channel %s: type=%s", channel, message.Type)
	return nil
}

// SubscribeToCourse subscribes to a course channel and returns a channel for messages
func (c *ChatPubSub) SubscribeToCourse(courseID uint64) (<-chan *ChatPubSubMessage, func(), error) {
	channel := c.GetCourseChannel(courseID)
	return c.subscribe(channel)
}

// SubscribeToUser subscribes to a user channel and returns a channel for messages
func (c *ChatPubSub) SubscribeToUser(userID uint64) (<-chan *ChatPubSubMessage, func(), error) {
	channel := c.GetUserChannel(userID)
	return c.subscribe(channel)
}

// subscribe subscribes to a Redis channel and returns a Go channel for messages
func (c *ChatPubSub) subscribe(channel string) (<-chan *ChatPubSubMessage, func(), error) {
	pubsub := c.client.Subscribe(c.ctx, channel)

	// Test the connection
	_, err := pubsub.Receive(c.ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	messageChan := make(chan *ChatPubSubMessage, 100) // Buffer for 100 messages

	// Start a goroutine to handle incoming messages
	go func() {
		defer close(messageChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			var chatMessage ChatPubSubMessage
			if err := json.Unmarshal([]byte(msg.Payload), &chatMessage); err != nil {
				log.Printf("Failed to unmarshal message from channel %s: %v", channel, err)
				continue
			}

			select {
			case messageChan <- &chatMessage:
			case <-c.ctx.Done():
				return
			default:
				log.Printf("Message buffer full for channel %s, dropping message", channel)
			}
		}
	}()

	// Return cleanup function
	cleanup := func() {
		pubsub.Close()
	}

	log.Printf("Subscribed to channel: %s", channel)
	return messageChan, cleanup, nil
}

// SubscribeToPrivateMessages subscribes to private messages between two users in a course
func (c *ChatPubSub) SubscribeToPrivateMessages(courseID, userID1, userID2 uint64) (<-chan *ChatPubSubMessage, func(), error) {
	channel := c.GetPrivateMessageChannel(courseID, userID1, userID2)
	return c.subscribe(channel)
}

func (c *ChatPubSub) SubscribeToAllPrivateMessages(courseID, userID uint64) (<-chan *ChatPubSubMessage, func(), error) {
	pattern1 := fmt.Sprintf("chat:private:course:%d:users:%d:*", courseID, userID)
	pattern2 := fmt.Sprintf("chat:private:course:%d:users:*:%d", courseID, userID)

	pubsub := c.client.PSubscribe(c.ctx, pattern1, pattern2)

	_, err := pubsub.Receive(c.ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to private message patterns: %w", err)
	}

	messageChan := make(chan *ChatPubSubMessage, 100)

	go func() {
		defer close(messageChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			var chatMessage ChatPubSubMessage
			if err := json.Unmarshal([]byte(msg.Payload), &chatMessage); err != nil {
				log.Printf("Failed to unmarshal message from channel %s: %v", msg.Channel, err)
				continue
			}

			if chatMessage.Data != nil {
				if chatMsgMap, ok := chatMessage.Data.(map[string]interface{}); ok {
					if recipientID, exists := chatMsgMap["recipient_id"]; exists {
						var rid uint64
						switch v := recipientID.(type) {
						case uint64:
							rid = v
						case float64:
							rid = uint64(v)
						case int:
							rid = uint64(v)
						}
						if rid != userID {
							continue
						}
					}
				}
			}

			select {
			case messageChan <- &chatMessage:
			case <-c.ctx.Done():
				return
			default:
				log.Printf("Message buffer full, dropping message from channel %s", msg.Channel)
			}
		}
	}()

	cleanup := func() {
		pubsub.Close()
	}

	log.Printf("Subscribed to all private messages for user %d in course %d", userID, courseID)
	return messageChan, cleanup, nil
}

// SubscribeToMultipleCourses subscribes to multiple course channels
func (c *ChatPubSub) SubscribeToMultipleCourses(courseIDs []uint64) (<-chan *ChatPubSubMessage, func(), error) {
	if len(courseIDs) == 0 {
		return nil, nil, fmt.Errorf("no course IDs provided")
	}

	channels := make([]string, len(courseIDs))
	for i, courseID := range courseIDs {
		channels[i] = c.GetCourseChannel(courseID)
	}

	pubsub := c.client.Subscribe(c.ctx, channels...)

	// Test the connection
	_, err := pubsub.Receive(c.ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to channels: %w", err)
	}

	messageChan := make(chan *ChatPubSubMessage, 100)

	go func() {
		defer close(messageChan)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for msg := range ch {
			var chatMessage ChatPubSubMessage
			if err := json.Unmarshal([]byte(msg.Payload), &chatMessage); err != nil {
				log.Printf("Failed to unmarshal message from channel %s: %v", msg.Channel, err)
				continue
			}

			select {
			case messageChan <- &chatMessage:
			case <-c.ctx.Done():
				return
			default:
				log.Printf("Message buffer full, dropping message from channel %s", msg.Channel)
			}
		}
	}()

	cleanup := func() {
		pubsub.Close()
	}

	log.Printf("Subscribed to %d course channels", len(courseIDs))
	return messageChan, cleanup, nil
}

// GetActiveChannels returns a list of active chat channels (for monitoring)
func (c *ChatPubSub) GetActiveChannels() ([]string, error) {
	channels, err := c.client.PubSubChannels(c.ctx, "chat:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}
	return channels, nil
}

// GetChannelSubscribers returns the number of subscribers for a channel
func (c *ChatPubSub) GetChannelSubscribers(channel string) (int64, error) {
	counts, err := c.client.PubSubNumSub(c.ctx, channel).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get subscriber count: %w", err)
	}

	if count, exists := counts[channel]; exists {
		return count, nil
	}
	return 0, nil
}

// Global instance
var chatPubSub *ChatPubSub

// GetChatPubSub returns the global ChatPubSub instance
func GetChatPubSub() *ChatPubSub {
	if chatPubSub == nil {
		chatPubSub = NewChatPubSub()
	}
	return chatPubSub
}

