package services

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebaseService struct {
	app             *firebase.App
	messagingClient *messaging.Client
	ctx             context.Context
}

var firebaseServiceInstance *FirebaseService

// InitFirebase khởi tạo Firebase App từ JSON credentials string
func InitFirebase(credentialsJSON string) (*FirebaseService, error) {
	if firebaseServiceInstance != nil {
		return firebaseServiceInstance, nil
	}

	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(credentialsJSON))

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %v", err)
	}

	firebaseServiceInstance = &FirebaseService{
		app:             app,
		messagingClient: messagingClient,
		ctx:             ctx,
	}

	log.Println("✅ Firebase initialized successfully")
	return firebaseServiceInstance, nil
}

// GetFirebaseService lấy instance đã khởi tạo
func GetFirebaseService() *FirebaseService {
	return firebaseServiceInstance
}

// SendToToken gửi push notification đến 1 device token
func (s *FirebaseService) SendToToken(token, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: badgeCount(1),
				},
			},
		},
	}

	response, err := s.messagingClient.Send(s.ctx, message)
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	log.Printf("✅ Successfully sent message: %s\n", response)
	return nil
}

// SendMulticast gửi push notification đến nhiều devices (tối đa 500)
func (s *FirebaseService) SendMulticast(tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	if len(tokens) > 500 {
		return nil, fmt.Errorf("maximum 500 tokens per multicast")
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: badgeCount(1),
				},
			},
		},
	}

	batchResponse, err := s.messagingClient.SendEachForMulticast(s.ctx, message)
	if err != nil {
		return nil, fmt.Errorf("error sending multicast: %v", err)
	}

	log.Printf("✅ Multicast sent: %d success, %d failure\n",
		batchResponse.SuccessCount, batchResponse.FailureCount)

	return batchResponse, nil
}

// SendMulticastInBatches gửi push đến nhiều tokens, tự động chia batch 500
func (s *FirebaseService) SendMulticastInBatches(tokens []string, title, body string, data map[string]string) (int, int, error) {
	totalSuccess := 0
	totalFailure := 0

	for i := 0; i < len(tokens); i += 500 {
		end := i + 500
		if end > len(tokens) {
			end = len(tokens)
		}

		batch := tokens[i:end]
		batchResponse, err := s.SendMulticast(batch, title, body, data)
		if err != nil {
			log.Printf("⚠️ Error sending batch %d-%d: %v\n", i, end, err)
			continue
		}

		totalSuccess += batchResponse.SuccessCount
		totalFailure += batchResponse.FailureCount
	}

	log.Printf("📊 Total multicast: %d success, %d failure\n", totalSuccess, totalFailure)
	return totalSuccess, totalFailure, nil
}

// SendToTopic gửi push notification đến một topic (dùng cho system/school broadcast)
func (s *FirebaseService) SendToTopic(topic, title, body string, data map[string]string) error {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					Badge: badgeCount(1),
				},
			},
		},
	}

	response, err := s.messagingClient.Send(s.ctx, message)
	if err != nil {
		return fmt.Errorf("error sending to topic %s: %v", topic, err)
	}

	log.Printf("✅ Successfully sent message to topic '%s': %s\n", topic, response)
	return nil
}

// SubscribeToTopics subscribe device token to multiple topics
// Used for web clients that cannot subscribe client-side
func (s *FirebaseService) SubscribeToTopics(token string, topics []string) error {
	if s.messagingClient == nil {
		return fmt.Errorf("firebase messaging client not initialized")
	}

	for _, topic := range topics {
		_, err := s.messagingClient.SubscribeToTopic(s.ctx, []string{token}, topic)
		if err != nil {
			log.Printf("❌ Failed to subscribe token to topic '%s': %v", topic, err)
			return fmt.Errorf("failed to subscribe to topic '%s': %v", topic, err)
		}
		log.Printf("✅ Subscribed device to topic: %s", topic)
	}

	return nil
}

// UnsubscribeFromTopics unsubscribe device token from multiple topics
func (s *FirebaseService) UnsubscribeFromTopics(token string, topics []string) error {
	if s.messagingClient == nil {
		return fmt.Errorf("firebase messaging client not initialized")
	}

	for _, topic := range topics {
		_, err := s.messagingClient.UnsubscribeFromTopic(s.ctx, []string{token}, topic)
		if err != nil {
			log.Printf("❌ Failed to unsubscribe token from topic '%s': %v", topic, err)
			return fmt.Errorf("failed to unsubscribe from topic '%s': %v", topic, err)
		}
		log.Printf("✅ Unsubscribed device from topic: %s", topic)
	}

	return nil
}

// Helper function for badge count
func badgeCount(count int) *int {
	return &count
}
