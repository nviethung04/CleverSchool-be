package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"fmt"
	"time"
)

const (
	NotificationTypeNotice = "notice"
)

type PushNotificationService struct {
	firebaseService     *FirebaseService
	userDeviceRepo      repositories.UserDeviceRepository
	notificationLogRepo repositories.NotificationLogRepository
	firebaseRepo        repositories.FirebaseRepository
}

func NewPushNotificationService() *PushNotificationService {
	return &PushNotificationService{
		firebaseService:     GetFirebaseService(),
		userDeviceRepo:      repositories.NewUserDeviceRepository(),
		notificationLogRepo: repositories.NewNotificationLogRepository(),
		firebaseRepo:        repositories.NewFirebaseRepository(),
	}
}

func (s *PushNotificationService) SendNoticeToTarget(notice *models.Notice, targetType string, targetValue int64) error {
	if s.firebaseService == nil {
		return fmt.Errorf("firebase service not initialized")
	}

	content := ""
	if notice.Content != nil {
		content = *notice.Content
	}

	data := map[string]string{
		"notice_id": fmt.Sprintf("%d", notice.ID),
		"type":      NotificationTypeNotice,
	}

	if targetType == models.NoticeTypeSystem {
		config.Log.Info("📤 Sending system push to topic 'all_system' for notice %d\n", notice.ID)

		allUserCount := s.userDeviceRepo.GetCountDevice()
		err := s.firebaseService.SendToTopic("all_system", notice.Title, content, data)

		var errorMessage string
		successCount := 0
		failureCount := 0
		if err != nil {
			errorMessage = err.Error()
			failureCount = int(allUserCount)
		} else {
			successCount = int(allUserCount)
		}

		s.logNotificationResults(notice.ID, successCount, failureCount, targetType, targetValue, errorMessage)

		if err != nil {
			return fmt.Errorf("failed to send to system topic: %v", err)
		}
		config.Log.Info("✅ System push notification sent successfully (estimated %d users)\n", successCount)
		return nil
	}

	if targetType == models.NoticeTypeSchool {
		exists, err := s.firebaseRepo.ValidateTarget(targetType, targetValue)
		if err != nil {
			return fmt.Errorf("failed to validate school: %v", err)
		}
		if !exists {
			return fmt.Errorf("school with ID %d does not exist", targetValue)
		}

		topic := fmt.Sprintf("school_%d", targetValue)
		config.Log.Info("📤 Sending school push to topic '%s' for notice %d\n", topic, notice.ID)

		userIDs, _ := s.firebaseRepo.GetUserIDsBySchool(targetValue)

		var userCount int64
		if len(userIDs) > 0 {
			s.userDeviceRepo.GetCountDeviceByUserIds(userIDs)
		}

		err = s.firebaseService.SendToTopic(topic, notice.Title, content, data)

		var errorMessage string
		successCount := 0
		failureCount := 0
		if err != nil {
			errorMessage = err.Error()
			failureCount = int(userCount)
		} else {
			successCount = int(userCount)
		}

		s.logNotificationResults(notice.ID, successCount, failureCount, targetType, targetValue, errorMessage)

		if err != nil {
			return fmt.Errorf("failed to send to school topic: %v", err)
		}
		config.Log.Info("✅ School push notification sent successfully (estimated %d users)\n", successCount)
		return nil
	}

	exists, err := s.firebaseRepo.ValidateTarget(targetType, targetValue)
	if err != nil {
		return fmt.Errorf("failed to validate target: %v", err)
	}
	if !exists {
		return fmt.Errorf("%s with ID %d does not exist", targetType, targetValue)
	}

	var userIDs []int64

	switch targetType {
	case models.NoticeTypeClass:
		userIDs, err = s.firebaseRepo.GetUserIDsByClass(targetValue)
	case models.NoticeTypeCourse:
		userIDs, err = s.firebaseRepo.GetUserIDsByCourse(targetValue)
	case models.NoticeTypeUser:
		userIDs = []int64{targetValue}
	default:
		return fmt.Errorf("invalid target type: %s", targetType)
	}

	if err != nil {
		return fmt.Errorf("failed to get user IDs: %v", err)
	}

	if len(userIDs) == 0 {
		return fmt.Errorf("no users found for %s ID %d", targetType, targetValue)
	}

	config.Log.Info("📤 Sending push to %d users (%s:%d) for notice %d\n", len(userIDs), targetType, targetValue, notice.ID)

	tokens, err := s.userDeviceRepo.GetActiveTokensByUserIDs(userIDs)
	if err != nil {
		return fmt.Errorf("failed to get FCM tokens: %v", err)
	}

	if len(tokens) == 0 {
		return fmt.Errorf("no active device tokens found for users in %s ID %d", targetType, targetValue)
	}

	config.Log.Info("📱 Found %d active device tokens\n", len(tokens))

	successCount, failureCount, err := s.firebaseService.SendMulticastInBatches(
		tokens,
		notice.Title,
		content,
		data,
	)

	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	} else {
		errorMessage = ""
	}

	s.logNotificationResults(notice.ID, successCount, failureCount, targetType, targetValue, errorMessage)

	if err != nil {
		return fmt.Errorf("error sending push: %v", err)
	}

	config.Log.Info("✅ Push notification sent: %d success, %d failure\n", successCount, failureCount)
	return nil
}

func (s *PushNotificationService) SendToUser(userID int64, title, body string, data map[string]string) error {
	if s.firebaseService == nil {
		return fmt.Errorf("firebase service not initialized")
	}

	devices, err := s.userDeviceRepo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user devices: %v", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("user has no active devices")
	}

	tokens := make([]string, 0)
	for _, device := range devices {
		tokens = append(tokens, device.DeviceToken)
	}

	_, err = s.firebaseService.SendMulticast(tokens, title, body, data)
	return err
}

func (s *PushNotificationService) logNotificationResults(noticeID int64, successCount, failureCount int, targetType string, targetValue int64, errorMessage string) {
	status := models.NotificationStatusSuccess
	if errorMessage != "" || failureCount > 0 {
		if successCount == 0 {
			status = models.NotificationStatusFailed
		} else {
			status = models.NotificationStatusSent
		}
	}

	log := &models.NotificationLog{
		NoticeID: noticeID,
		Data: models.NotificationData{
			Type:         targetType,
			DataValue:    fmt.Sprintf("%d", targetValue),
			SuccessCount: int32(successCount),
			FailedCount:  int32(failureCount),
			Error:        errorMessage,
		},
		Status: status,
		SentAt: time.Now(),
	}

	if err := s.notificationLogRepo.CreateLog(log); err != nil {
		config.Log.Errorf("⚠️ Failed to create notification log for notice %d: %v", noticeID, err)
	} else {
		config.Log.Infof("📝 Logged notification result - NoticeID: %d, Type: %s, Success: %d, Failed: %d, Status: %s",
			noticeID, targetType, successCount, failureCount, status)
	}
}

func (s *PushNotificationService) RegisterDevice(userID int64, deviceToken, deviceType, platform, appVersion string) error {
	err := s.userDeviceRepo.UpsertDevice(userID, deviceToken, deviceType, platform, appVersion)
	if err != nil {
		return fmt.Errorf("failed to save device: %v", err)
	}

	schoolIDs, err := s.firebaseRepo.GetSchoolIDsByUserID(userID)
	if err != nil {
		config.Log.Info("⚠️ Failed to get user's school_ids: %v", err)
		return nil
	}

	topics := []string{"all_system"}
	for _, schoolID := range schoolIDs {
		if schoolID > 0 {
			topics = append(topics, fmt.Sprintf("school_%d", schoolID))
		}
	}

	if s.firebaseService != nil {
		err = s.firebaseService.SubscribeToTopics(deviceToken, topics)
		if err != nil {
			config.Log.Info("⚠️ Failed to subscribe device to topics: %v", err)
		} else {
			config.Log.Info("📱 Device subscribed to topics: %v", topics)
		}
	}

	return nil
}

func (s *PushNotificationService) UnregisterDevice(deviceToken string) error {
	device, err := s.userDeviceRepo.FindByDeviceToken(deviceToken, true)
	if err != nil {
		config.Log.Info("⚠️ Device not found or already inactive: %v", err)
		return s.userDeviceRepo.DeactivateDevice(deviceToken)
	}

	schoolIDs, err := s.firebaseRepo.GetSchoolIDsByUserID(device.UserID)
	if err != nil {
		config.Log.Info("⚠️ Failed to get user's school_ids: %v", err)
	}

	topics := []string{"all_system"}
	for _, schoolID := range schoolIDs {
		if schoolID > 0 {
			topics = append(topics, fmt.Sprintf("school_%d", schoolID))
		}
	}

	if s.firebaseService != nil {
		err = s.firebaseService.UnsubscribeFromTopics(deviceToken, topics)
		if err != nil {
			config.Log.Info("⚠️ Failed to unsubscribe device from topics: %v", err)
		} else {
			config.Log.Info("📱 Device unsubscribed from topics: %v", topics)
		}
	}

	return s.userDeviceRepo.DeactivateDevice(deviceToken)
}

func (s *PushNotificationService) GetUserDevices(userID int64) ([]models.UserDevice, error) {
	return s.userDeviceRepo.FindByUserID(userID)
}

func (s *PushNotificationService) GetNoticeByID(noticeID int64) (*models.Notice, error) {
	noticeRepo := repositories.NewNoticeRepository()
	return noticeRepo.FindByID(int(noticeID))
}
