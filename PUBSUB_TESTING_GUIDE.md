# 🧪 Redis Pub/Sub Testing Guide

## 🎯 **Hướng dẫn Test Pub/Sub**

### **1. Test Basic Connection**

Tạo file test đơn giản để kiểm tra Redis Pub/Sub:

```go
// test_pubsub_basic.go
package main

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/redis"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("🧪 Testing Redis Pub/Sub...")

	// Load config
	cfg := config.LoadConfig()
	
	// Initialize database (for Redis client)
	db.InitDB(cfg.DBMasterURL, cfg.DBSlaveURL)
	
	// Create Pub/Sub client
	pubsub := redis.NewChatPubSub()
	
	// Test 1: Check Redis connection
	fmt.Println("📡 Testing Redis connection...")
	ctx := context.Background()
	
	// Test ping
	pong, err := db.RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("❌ Redis connection failed:", err)
	}
	fmt.Printf("✅ Redis ping: %s\n", pong)
	
	// Test 2: Subscribe to channel
	fmt.Println("🔔 Testing subscription...")
	courseId := 123
	
	// Subscribe to course channel
	subscription, err := pubsub.SubscribeToCourse(ctx, courseId)
	if err != nil {
		log.Fatal("❌ Subscribe failed:", err)
	}
	defer subscription.Close()
	
	fmt.Printf("✅ Subscribed to course %d\n", courseId)
	
	// Test 3: Publish test message
	fmt.Println("📤 Publishing test message...")
	
	testMessage := redis.ChatNotification{
		Type:      "message",
		CourseID:  courseId,
		MessageID: 999,
		UserID:    1,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"content":    "Test message from Pub/Sub",
			"user_name":  "Test User",
			"created_at": time.Now().Format(time.RFC3339),
		},
	}
	
	// Publish message
	err = pubsub.PublishCourseMessage(ctx, courseId, testMessage)
	if err != nil {
		log.Fatal("❌ Publish failed:", err)
	}
	
	fmt.Printf("✅ Published message to course %d\n", courseId)
	
	// Test 4: Receive message
	fmt.Println("📥 Waiting for messages...")
	
	// Set timeout for receiving
	timeout := time.After(5 * time.Second)
	
	select {
	case msg := <-subscription.Channel():
		fmt.Printf("✅ Received message: %s\n", msg.Payload)
		
		// Parse the message
		var notification redis.ChatNotification
		if err := json.Unmarshal([]byte(msg.Payload), &notification); err == nil {
			fmt.Printf("   Type: %s\n", notification.Type)
			fmt.Printf("   Course ID: %d\n", notification.CourseID)
			fmt.Printf("   Message ID: %d\n", notification.MessageID)
			fmt.Printf("   User ID: %d\n", notification.UserID)
		}
		
	case <-timeout:
		fmt.Println("⏰ Timeout waiting for message")
	}
	
	fmt.Println("✅ Pub/Sub test completed!")
}
```

### **2. Test Multiple Channels**

```go
// test_pubsub_multiple.go
package main

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("🧪 Testing Multiple Channel Pub/Sub...")

	// Setup
	cfg := config.LoadConfig()
	db.InitDB(cfg.DBMasterURL, cfg.DBSlaveURL)
	pubsub := redis.NewChatPubSub()
	ctx := context.Background()

	// Test multiple courses
	courseIds := []int{123, 456, 789}
	var wg sync.WaitGroup

	// Subscribe to multiple courses
	for _, courseId := range courseIds {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			subscription, err := pubsub.SubscribeToCourse(ctx, id)
			if err != nil {
				log.Printf("❌ Subscribe to course %d failed: %v", id, err)
				return
			}
			defer subscription.Close()
			
			fmt.Printf("✅ Subscribed to course %d\n", id)
			
			// Listen for messages
			timeout := time.After(10 * time.Second)
			
			select {
			case msg := <-subscription.Channel():
				var notification redis.ChatNotification
				if err := json.Unmarshal([]byte(msg.Payload), &notification); err == nil {
					fmt.Printf("📨 Course %d received: %s from user %d\n", 
						notification.CourseID, notification.Type, notification.UserID)
				}
			case <-timeout:
				fmt.Printf("⏰ Course %d timeout\n", id)
			}
		}(courseId)
	}

	// Wait a bit for subscriptions to be ready
	time.Sleep(1 * time.Second)

	// Publish to each course
	for _, courseId := range courseIds {
		testMessage := redis.ChatNotification{
			Type:      "message",
			CourseID:  courseId,
			MessageID: 1000 + courseId,
			UserID:    1,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"content": fmt.Sprintf("Test message for course %d", courseId),
			},
		}

		err := pubsub.PublishCourseMessage(ctx, courseId, testMessage)
		if err != nil {
			log.Printf("❌ Publish to course %d failed: %v", courseId, err)
		} else {
			fmt.Printf("📤 Published to course %d\n", courseId)
		}
	}

	// Wait for all goroutines
	wg.Wait()
	fmt.Println("✅ Multiple channel test completed!")
}
```

### **3. Test SSE Integration**

```go
// test_sse_integration.go
package main

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/redis"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 Testing SSE Integration...")

	// Setup
	cfg := config.LoadConfig()
	db.InitDB(cfg.DBMasterURL, cfg.DBSlaveURL)
	
	// Start test HTTP server in background
	go func() {
		fmt.Println("🌐 Starting test HTTP server on :8081...")
		
		http.HandleFunc("/test-sse", func(w http.ResponseWriter, r *http.Request) {
			// Set SSE headers
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			// Create Pub/Sub client
			pubsub := redis.NewChatPubSub()
			ctx := r.Context()

			// Subscribe to course 123
			subscription, err := pubsub.SubscribeToCourse(ctx, 123)
			if err != nil {
				fmt.Fprintf(w, "data: {\"error\": \"Failed to subscribe\"}\n\n")
				return
			}
			defer subscription.Close()

			// Send connection confirmation
			fmt.Fprintf(w, "event: connected\n")
			fmt.Fprintf(w, "data: {\"status\": \"connected\", \"course_id\": 123}\n\n")
			w.(http.Flusher).Flush()

			// Listen for messages
			for {
				select {
				case msg := <-subscription.Channel():
					// Forward message to SSE
					fmt.Fprintf(w, "event: message\n")
					fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
					w.(http.Flusher).Flush()

				case <-ctx.Done():
					fmt.Println("🔌 SSE client disconnected")
					return
				}
			}
		})

		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test publishing messages
	pubsub := redis.NewChatPubSub()
	ctx := context.Background()

	fmt.Println("📤 Publishing test messages...")

	for i := 1; i <= 5; i++ {
		testMessage := redis.ChatNotification{
			Type:      "message",
			CourseID:  123,
			MessageID: i,
			UserID:    1,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"content":   fmt.Sprintf("Test message #%d", i),
				"user_name": "Test User",
			},
		}

		err := pubsub.PublishCourseMessage(ctx, 123, testMessage)
		if err != nil {
			log.Printf("❌ Publish failed: %v", err)
		} else {
			fmt.Printf("✅ Published message #%d\n", i)
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("🌐 Test SSE at: http://localhost:8081/test-sse")
	fmt.Println("⏰ Keeping server running for 30 seconds...")
	time.Sleep(30 * time.Second)
}
```

### **4. Test Performance**

```go
// test_pubsub_performance.go
package main

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/redis"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("🚀 Testing Pub/Sub Performance...")

	cfg := config.LoadConfig()
	db.InitDB(cfg.DBMasterURL, cfg.DBSlaveURL)
	pubsub := redis.NewChatPubSub()
	ctx := context.Background()

	// Test parameters
	numMessages := 1000
	numSubscribers := 10
	courseId := 123

	// Create multiple subscribers
	var wg sync.WaitGroup
	receivedCount := 0
	var mutex sync.Mutex

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(subscriberId int) {
			defer wg.Done()

			subscription, err := pubsub.SubscribeToCourse(ctx, courseId)
			if err != nil {
				log.Printf("❌ Subscriber %d failed: %v", subscriberId, err)
				return
			}
			defer subscription.Close()

			localCount := 0
			timeout := time.After(30 * time.Second)

			for {
				select {
				case <-subscription.Channel():
					localCount++
					mutex.Lock()
					receivedCount++
					mutex.Unlock()

				case <-timeout:
					fmt.Printf("📊 Subscriber %d received %d messages\n", subscriberId, localCount)
					return
				}
			}
		}(i)
	}

	// Wait for subscribers to be ready
	time.Sleep(2 * time.Second)

	// Start publishing
	fmt.Printf("📤 Publishing %d messages...\n", numMessages)
	startTime := time.Now()

	for i := 0; i < numMessages; i++ {
		testMessage := redis.ChatNotification{
			Type:      "message",
			CourseID:  courseId,
			MessageID: i,
			UserID:    1,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"content": fmt.Sprintf("Performance test message #%d", i),
			},
		}

		err := pubsub.PublishCourseMessage(ctx, courseId, testMessage)
		if err != nil {
			log.Printf("❌ Publish %d failed: %v", i, err)
		}

		// Small delay to avoid overwhelming
		if i%100 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	publishDuration := time.Since(startTime)
	fmt.Printf("✅ Published %d messages in %v\n", numMessages, publishDuration)
	fmt.Printf("📊 Rate: %.2f messages/second\n", float64(numMessages)/publishDuration.Seconds())

	// Wait for messages to be received
	time.Sleep(5 * time.Second)

	mutex.Lock()
	totalReceived := receivedCount
	mutex.Unlock()

	fmt.Printf("📊 Performance Results:\n")
	fmt.Printf("   Messages sent: %d\n", numMessages)
	fmt.Printf("   Subscribers: %d\n", numSubscribers)
	fmt.Printf("   Total received: %d\n", totalReceived)
	fmt.Printf("   Expected: %d\n", numMessages*numSubscribers)
	fmt.Printf("   Success rate: %.2f%%\n", float64(totalReceived)/float64(numMessages*numSubscribers)*100)

	// Cancel context to stop subscribers
	time.Sleep(2 * time.Second)
}
```

## 🚀 **Cách chạy tests:**

### **Bước 1: Test cơ bản**
```bash
go run test_pubsub_basic.go
```

### **Bước 2: Test multiple channels**
```bash
go run test_pubsub_multiple.go
```

### **Bước 3: Test SSE integration**
```bash
go run test_sse_integration.go
```

### **Bước 4: Test performance**
```bash
go run test_pubsub_performance.go
```

## 📋 **Expected Results:**

✅ **Basic Test:** Connection, publish, receive
✅ **Multiple Test:** Multiple course channels working
✅ **SSE Test:** Real-time web integration
✅ **Performance Test:** High throughput capabilities

## 🔧 **Troubleshooting:**

- ❌ **Redis connection failed** → Check Redis server running
- ❌ **Subscribe failed** → Check Redis permissions
- ❌ **Timeout** → Check network/Redis performance
- ❌ **Performance low** → Check Redis memory/CPU

**Ready to test! 🧪**
