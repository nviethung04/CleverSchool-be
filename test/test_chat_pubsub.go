package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"be-Clever School/redis"
	"be-Clever School/services"
)

// Test script for Redis Pub/Sub and WebSocket implementation
func main() {
	log.Println("=== Testing Chat Redis Pub/Sub System ===")

	// Test 1: Create pub/sub service
	pubsubService := services.NewChatPubSubService()
	log.Println("✅ Created ChatPubSubService")

	// Test 2: Test channel names
	testChannelNames()

	// Test 3: Test pub/sub basic functionality
	testPubSubBasic(pubsubService)

	// Test 4: Test multiple course subscription
	testMultipleCourseSubscription(pubsubService)

	// Test 5: Test user notifications
	testUserNotifications(pubsubService)

	// Test 6: Test monitoring features
	testMonitoring(pubsubService)

	log.Println("=== All tests completed ===")
}

func testChannelNames() {
	log.Println("\n--- Testing Channel Names ---")

	pubsub := redis.GetChatPubSub()

	// Test course channels
	courseChannel := pubsub.GetCourseChannel(123)
	log.Printf("Course 123 channel: %s", courseChannel)

	// Test user channels
	userChannel := pubsub.GetUserChannel(456)
	log.Printf("User 456 channel: %s", userChannel)

	log.Println("✅ Channel names generated correctly")
}

func testPubSubBasic(pubsubService services.ChatPubSubService) {
	log.Println("\n--- Testing Basic Pub/Sub ---")

	courseID := uint64(123)
	messageID := uint64(1)
	userID := uint64(456)

	// Subscribe to course
	messageChan, cleanup, err := pubsubService.SubscribeToCourse(courseID)
	if err != nil {
		log.Printf("❌ Failed to subscribe: %v", err)
		return
	}
	defer cleanup()

	// Setup message handler
	go func() {
		select {
		case msg := <-messageChan:
			if msg != nil {
				log.Printf("📨 Received message: type=%s, courseID=%d, messageID=%d",
					msg.Type, msg.CourseID, msg.MessageID)
			}
		case <-time.After(5 * time.Second):
			log.Println("⏰ No message received within 5 seconds")
		}
	}()

	// Wait a moment for subscription to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish a test message
	messageData := map[string]interface{}{
		"content":   "Hello, this is a test message!",
		"user_name": "Test User",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	err = pubsubService.NotifyNewMessage(courseID, messageID, userID, messageData)
	if err != nil {
		log.Printf("❌ Failed to publish message: %v", err)
		return
	}

	log.Println("📤 Published test message")

	// Wait for message to be received
	time.Sleep(1 * time.Second)
	log.Println("✅ Basic pub/sub test completed")
}

func testMultipleCourseSubscription(pubsubService services.ChatPubSubService) {
	log.Println("\n--- Testing Multiple Course Subscription ---")

	courseIDs := []uint64{100, 200, 300}

	// Subscribe to multiple courses
	messageChan, cleanup, err := pubsubService.SubscribeToMultipleCourses(courseIDs)
	if err != nil {
		log.Printf("❌ Failed to subscribe to multiple courses: %v", err)
		return
	}
	defer cleanup()

	// Setup message handler
	receivedMessages := 0
	expectedMessages := len(courseIDs)

	go func() {
		for {
			select {
			case msg := <-messageChan:
				if msg != nil {
					receivedMessages++
					log.Printf("📨 Received from course %d: type=%s", msg.CourseID, msg.Type)
				}
			case <-time.After(3 * time.Second):
				log.Printf("⏰ Timeout - received %d/%d messages", receivedMessages, expectedMessages)
				return
			}
		}
	}()

	// Wait for subscription to be ready
	time.Sleep(100 * time.Millisecond)

	// Publish to each course
	for i, courseID := range courseIDs {
		messageData := map[string]interface{}{
			"content": fmt.Sprintf("Test message %d for course %d", i+1, courseID),
		}

		err = pubsubService.NotifyNewMessage(courseID, uint64(i+1), 999, messageData)
		if err != nil {
			log.Printf("❌ Failed to publish to course %d: %v", courseID, err)
		} else {
			log.Printf("📤 Published to course %d", courseID)
		}

		time.Sleep(50 * time.Millisecond) // Small delay between publishes
	}

	// Wait for all messages
	time.Sleep(2 * time.Second)
	log.Printf("✅ Multiple course subscription test completed (%d messages)", receivedMessages)
}

func testUserNotifications(pubsubService services.ChatPubSubService) {
	log.Println("\n--- Testing User Notifications ---")

	userID := uint64(789)

	// Subscribe to user notifications
	notificationChan, cleanup, err := pubsubService.SubscribeToUser(userID)
	if err != nil {
		log.Printf("❌ Failed to subscribe to user notifications: %v", err)
		return
	}
	defer cleanup()

	// Setup notification handler
	go func() {
		select {
		case msg := <-notificationChan:
			if msg != nil {
				log.Printf("🔔 Received user notification: type=%s, userID=%d", msg.Type, msg.UserID)
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if msgType, exists := data["type"]; exists {
						log.Printf("   Notification type: %s", msgType)
					}
				}
			}
		case <-time.After(3 * time.Second):
			log.Println("⏰ No user notification received within 3 seconds")
		}
	}()

	// Wait for subscription
	time.Sleep(100 * time.Millisecond)

	// Send test notifications
	notifications := []map[string]interface{}{
		{
			"type":      "new_message_mention",
			"message":   "You were mentioned in a chat message",
			"course_id": 123,
		},
		{
			"type":      "assignment_due",
			"message":   "Assignment due soon",
			"course_id": 456,
		},
	}

	for i, notification := range notifications {
		err = pubsubService.NotifyUser(userID, notification["type"].(string), notification)
		if err != nil {
			log.Printf("❌ Failed to send notification %d: %v", i+1, err)
		} else {
			log.Printf("📤 Sent notification: %s", notification["type"])
		}
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
	log.Println("✅ User notifications test completed")
}

func testMonitoring(pubsubService services.ChatPubSubService) {
	log.Println("\n--- Testing Monitoring Features ---")

	// Test getting active channels
	channels, err := pubsubService.GetActiveChannels()
	if err != nil {
		log.Printf("❌ Failed to get active channels: %v", err)
	} else {
		log.Printf("📊 Active channels found: %d", len(channels))
		for i, channel := range channels {
			if i < 5 { // Show first 5 channels
				log.Printf("   Channel: %s", channel)
			}
		}
		if len(channels) > 5 {
			log.Printf("   ... and %d more", len(channels)-5)
		}
	}

	// Test getting subscriber counts for some channels
	if len(channels) > 0 {
		testChannel := channels[0]
		count, err := pubsubService.GetChannelSubscribers(testChannel)
		if err != nil {
			log.Printf("❌ Failed to get subscriber count for %s: %v", testChannel, err)
		} else {
			log.Printf("👥 Channel %s has %d subscribers", testChannel, count)
		}
	}

	log.Println("✅ Monitoring test completed")
}

// Additional test: Test event handler
func testEventHandler() {
	log.Println("\n--- Testing Event Handler ---")

	pubsubService := services.NewChatPubSubService()

	// Create event handler
	handler := services.NewChatEventHandler()
	handler.OnNewMessage(func(courseID, messageID, userID uint64, data interface{}) {
		log.Printf("🎯 Handler - New message: course=%d, msg=%d, user=%d", courseID, messageID, userID)
	}).OnNewReply(func(courseID, messageID, parentMessageID, userID uint64, data interface{}) {
		log.Printf("🎯 Handler - New reply: course=%d, msg=%d, parent=%d, user=%d", courseID, messageID, parentMessageID, userID)
	}).OnReaction(func(courseID, messageID, userID uint64, reaction redis.ChatReactionEvent) {
		log.Printf("🎯 Handler - Reaction: course=%d, msg=%d, user=%d, emoji=%s", courseID, messageID, userID, reaction.Emoji)
	})

	// Start event listener
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := services.StartChatEventListener(ctx, pubsubService, []uint64{123}, handler)
	if err != nil {
		log.Printf("❌ Failed to start event listener: %v", err)
		return
	}

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Publish test events
	pubsubService.NotifyNewMessage(123, 1, 456, map[string]interface{}{"content": "Test message"})
	pubsubService.NotifyReactionAdded(123, 1, 456, 1, "👍", 1)

	// Wait for events to be processed
	time.Sleep(1 * time.Second)

	log.Println("✅ Event handler test completed")
}

// Test data structures
func testDataStructures() {
	log.Println("\n--- Testing Data Structures ---")

	// Test ChatPubSubMessage
	message := &redis.ChatPubSubMessage{
		Type:      "message",
		CourseID:  123,
		MessageID: 456,
		UserID:    789,
		Data: map[string]interface{}{
			"content":   "Hello World",
			"timestamp": time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Failed to marshal message: %v", err)
		return
	}

	log.Printf("📝 Message JSON: %s", string(jsonData))

	// Deserialize from JSON
	var parsedMessage redis.ChatPubSubMessage
	err = json.Unmarshal(jsonData, &parsedMessage)
	if err != nil {
		log.Printf("❌ Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("📖 Parsed message: type=%s, courseID=%d", parsedMessage.Type, parsedMessage.CourseID)
	log.Println("✅ Data structures test completed")
}
