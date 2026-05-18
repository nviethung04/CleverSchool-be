package services

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ChatMessageService interface {
	SendMessage(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	SendMessageWithFiles(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest, files []*multipart.FileHeader) (*dto.ChatMessageResponse, error)
	SendMessageWithMedias(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	GetMessages(c *gin.Context, courseId uint64, page, limit int) (*dto.ChatMessagesListResponse, error)
	DeleteMessage(c *gin.Context, messageId uint64, userId uint64) error
	TogglePinMessage(c *gin.Context, messageId uint64, userId uint64) error
	GetPinnedMessages(c *gin.Context, courseId uint64) ([]dto.ChatMessageResponse, error)
	CountMessages(c *gin.Context, courseId uint64) (int64, error)
	GetRecentMessages(c *gin.Context, courseId uint64, limit int) ([]dto.ChatMessageResponse, error)

	SendToRecipient(c *gin.Context, courseId, senderId, recipientId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	GetMessagesWithRecipient(c *gin.Context, courseId, senderId, recipientId uint64, page, limit int) (*dto.ChatMessagesListResponse, error)
	GetRecentSenders(c *gin.Context, courseId, currentUserId uint64, limit int) ([]dto.RecentSenderResponse, error)
}

type chatMessageService struct {
	repo             repositories.ChatMessageRepository
	userRepo         repositories.UserRepository
	courseRepo       repositories.CourseRepository
	messageMediaRepo repositories.MessageMediaRepository
	mediaRepo        repositories.MediaRepository
	pubsubService    ChatPubSubService
}

func NewChatMessageService(
	repo repositories.ChatMessageRepository,
	userRepo repositories.UserRepository,
	courseRepo repositories.CourseRepository,
	messageMediaRepo repositories.MessageMediaRepository,
	mediaRepo repositories.MediaRepository,
) ChatMessageService {
	return &chatMessageService{
		repo:             repo,
		userRepo:         userRepo,
		courseRepo:       courseRepo,
		messageMediaRepo: messageMediaRepo,
		mediaRepo:        mediaRepo,
		pubsubService:    NewChatPubSubService(),
	}
}

func (s *chatMessageService) determineMessageType(content string, files []*multipart.FileHeader) string {
	if strings.TrimSpace(content) != "" {
		return "text"
	}

	if len(files) == 0 {
		return "text"
	}

	if len(files) == 1 {
		filename := strings.ToLower(files[0].Filename)
		if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") ||
			strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".gif") ||
			strings.HasSuffix(filename, ".webp") {
			return "image"
		}
		if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".avi") ||
			strings.HasSuffix(filename, ".mov") || strings.HasSuffix(filename, ".wmv") {
			return "video"
		}
		if strings.HasSuffix(filename, ".mp3") || strings.HasSuffix(filename, ".wav") ||
			strings.HasSuffix(filename, ".flac") || strings.HasSuffix(filename, ".aac") {
			return "audio"
		}
		return "file"
	}

	var fileTypes []string
	for _, file := range files {
		filename := strings.ToLower(file.Filename)
		if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") ||
			strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".gif") ||
			strings.HasSuffix(filename, ".webp") {
			fileTypes = append(fileTypes, "image")
		} else if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".avi") ||
			strings.HasSuffix(filename, ".mov") || strings.HasSuffix(filename, ".wmv") {
			fileTypes = append(fileTypes, "video")
		} else if strings.HasSuffix(filename, ".mp3") || strings.HasSuffix(filename, ".wav") ||
			strings.HasSuffix(filename, ".flac") || strings.HasSuffix(filename, ".aac") {
			fileTypes = append(fileTypes, "audio")
		} else {
			fileTypes = append(fileTypes, "file")
		}
	}

	if len(fileTypes) > 0 {
		firstType := fileTypes[0]
		allSameType := true
		for _, t := range fileTypes {
			if t != firstType {
				allSameType = false
				break
			}
		}
		if allSameType {
			return firstType
		}
	}

	return "mixed"
}

func (s *chatMessageService) SendMessage(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	s.userRepo.SetContext(c)
	user, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if len(req.Content) == 0 {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	messageType := "text"

	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType,
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	response := &dto.ChatMessageResponse{
		ID:               message.ID,
		CourseID:         courseId,
		UserID:           userId,
		Content:          message.Content,
		MessageType:      message.MessageType,
		IsPinned:         message.IsPinned,
		IsEdited:         message.IsEdited,
		EditedAt:         message.EditedAt,
		ReplyToMessageID: message.ReplyToMessageID,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
		User: dto.ChatUserResponse{
			ID:       uint64(user.ID),
			Username: user.Username,
			FullName: user.Name,
			Avatar:   s.generateAvatarURL(user.AvatarInfo),
		},
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewMessage: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
			config.Log.Warn("⚠️ Failed to publish new message notification: %v", err)
		}
	}()

	return response, nil
}
func (s *chatMessageService) SendMessageWithMedias(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	s.userRepo.SetContext(c)
	user, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		config.Log.Errorf("User not found in SendMessageWithMedias: userId=%d, error=%v", userId, err)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("message must have either content or media attachments")
	}
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	var medias []*models.Media
	if len(req.MediaIDs) > 0 {
		medias = make([]*models.Media, 0, len(req.MediaIDs))
		for _, mediaID := range req.MediaIDs {
			if mediaID <= 0 {
				config.Log.Warnf("Invalid media ID in request: %d", mediaID)
				return nil, fmt.Errorf("invalid media ID: %d", mediaID)
			}
			
			m, err := s.mediaRepo.FindById(mediaID)
			if err != nil {
				config.Log.Errorf("Media not found in SendMessageWithMedias: mediaID=%d, error=%v", mediaID, err)
				return nil, fmt.Errorf("media with ID %d not found: %w", mediaID, err)
			}
			if m == nil {
				config.Log.Errorf("Media is nil for ID: %d", mediaID)
				return nil, fmt.Errorf("media with ID %d not found", mediaID)
			}
			medias = append(medias, m)
		}
	}

	messageType := "text"
	if len(req.MediaIDs) > 0 {
		messageType = s.determineMessageTypeFromMedias(ctx, req.MediaIDs)
		if messageType == "" {
			messageType = "file" // Default fallback
		}
	}

	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType,
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateWithMedias(ctx, message, req.MediaIDs); err != nil {
		config.Log.Errorf("Failed to create message with medias: courseId=%d, userId=%d, mediaIDs=%v, error=%v", courseId, userId, req.MediaIDs, err)
		return nil, fmt.Errorf("failed to create message with medias: %w", err)
	}

	response := &dto.ChatMessageResponse{
		ID:               message.ID,
		CourseID:         courseId,
		UserID:           userId,
		Content:          message.Content,
		MessageType:      message.MessageType,
		IsPinned:         message.IsPinned,
		IsEdited:         message.IsEdited,
		EditedAt:         message.EditedAt,
		ReplyToMessageID: message.ReplyToMessageID,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
		User: dto.ChatUserResponse{
			ID:       uint64(user.ID),
			Username: user.Username,
			FullName: user.Name,
			Avatar:   s.generateAvatarURL(user.AvatarInfo),
		},
	}

	if len(medias) > 0 {
		response.Medias = make([]dto.ChatMediaResponse, len(medias))
		for i, media := range medias {
			var fullStaticURL *string
			if media.StaticURL != nil {
				diskName := "s3"
				if media.DiskName != nil {
					diskName = *media.DiskName
				}
				fullURL := utils.StaticURL(*media.StaticURL, diskName)
				fullStaticURL = &fullURL
			}

			response.Medias[i] = dto.ChatMediaResponse{
				ID:            media.ID,
				FileName:      media.FileName,
				FilePath:      media.FilePath,
				FullPath:      media.FullPath,
				FileType:      media.FileType,
				FileSize:      media.FileSize,
				FileExtension: media.FileExtension,
				DiskName:      media.DiskName,
				StaticURL:     fullStaticURL,
				SortOrder:     i,
			}
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewMessage: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
			config.Log.Warn("⚠️ Failed to publish new message notification: %v", err)
		}
	}()

	return response, nil
}

func (s *chatMessageService) SendMessageWithFiles(c *gin.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest, files []*multipart.FileHeader) (*dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	s.userRepo.SetContext(c)
	user, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	messageType := s.determineMessageType(req.Content, files)

	validTypes := map[string]bool{
		"text":  true,
		"file":  true,
		"image": true,
		"video": true,
		"audio": true,
		"mixed": true,
	}
	if !validTypes[messageType] {
		return nil, fmt.Errorf("invalid message type: %s", messageType)
	}

	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	if len(req.Content) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("message must have either content or files")
	}

	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType,
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	response := &dto.ChatMessageResponse{
		ID:               message.ID,
		CourseID:         courseId,
		UserID:           userId,
		Content:          message.Content,
		MessageType:      message.MessageType,
		IsPinned:         message.IsPinned,
		IsEdited:         message.IsEdited,
		EditedAt:         message.EditedAt,
		ReplyToMessageID: message.ReplyToMessageID,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
		User: dto.ChatUserResponse{
			ID:       uint64(user.ID),
			Username: user.Username,
			FullName: user.Name,
			Avatar:   s.generateAvatarURL(user.AvatarInfo),
		},
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewMessage: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
			config.Log.Warn("⚠️ Failed to publish new message notification: %v", err)
		}
	}()

	return response, nil
}

func (s *chatMessageService) GetMessages(c *gin.Context, courseId uint64, page, limit int) (*dto.ChatMessagesListResponse, error) {
	ctx := c.Request.Context()
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	messages, total, err := s.repo.GetByCourseID(ctx, courseId, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	responseMessages := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responseMessages[i] = *s.mapToResponse(&msg)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &dto.ChatMessagesListResponse{
		Messages: responseMessages,
		Pagination: dto.PaginationResponse{
			Page:      page,
			Limit:     limit,
			Total:     int(total),
			TotalPage: totalPages,
			HasMore:   page < totalPages,
		},
	}, nil
}

func (s *chatMessageService) DeleteMessage(c *gin.Context, messageId uint64, userId uint64) error {
	ctx := c.Request.Context()
	message, err := s.repo.GetByID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	if message.UserID != userId {
		return fmt.Errorf("unauthorized: only message owner can delete this message")
	}

	if time.Since(message.CreatedAt) > 24*time.Hour {
		return fmt.Errorf("cannot delete message older than 24 hours")
	}

	err = s.repo.DeleteRepliesByID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("failed to delete message replies: %w", err)
	}

	err = s.repo.Delete(ctx, messageId)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyMessageDeleted: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyMessageDeleted(message.CourseID, messageId, userId); err != nil {
			config.Log.Warn("⚠️ Failed to publish message deleted notification: %v", err)
		}
	}()

	return nil
}

func (s *chatMessageService) TogglePinMessage(c *gin.Context, messageId uint64, userId uint64) error {
	ctx := c.Request.Context()
	message, err := s.repo.GetByID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	wasPinned := message.IsPinned

	err = s.repo.TogglePin(ctx, messageId)
	if err != nil {
		return err
	}

	newPinStatus := !wasPinned
	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyMessagePinned: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyMessagePinned(message.CourseID, messageId, userId, newPinStatus); err != nil {
			config.Log.Warn("⚠️ Failed to publish message pin notification: %v", err)
		}
	}()

	return nil
}

func (s *chatMessageService) GetPinnedMessages(c *gin.Context, courseId uint64) ([]dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	messages, err := s.repo.GetPinnedByCourseID(ctx, courseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get pinned messages: %w", err)
	}

	responses := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.mapToResponse(&msg)
	}

	return responses, nil
}

func (s *chatMessageService) CountMessages(c *gin.Context, courseId uint64) (int64, error) {
	ctx := c.Request.Context()
	return s.repo.CountByCourseID(ctx, courseId)
}

func (s *chatMessageService) GetRecentMessages(c *gin.Context, courseId uint64, limit int) ([]dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	if limit < 1 || limit > 50 {
		limit = 10
	}

	messages, err := s.repo.GetRecentByCourseID(ctx, courseId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}

	responses := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.mapToResponse(&msg)
	}

	return responses, nil
}

func (s *chatMessageService) mapToResponse(message *models.ChatMessage) *dto.ChatMessageResponse {
	var recipient *dto.ChatUserResponse

	if message.Recipient != nil {
		recipient = &dto.ChatUserResponse{
			ID:       uint64(message.Recipient.ID),
			Username: message.Recipient.Username,
			FullName: message.Recipient.Name,
			Avatar:   s.generateAvatarURL(message.Recipient.AvatarInfo),
		}
	}

	response := &dto.ChatMessageResponse{
		ID:               message.ID,
		CourseID:         message.CourseID,
		UserID:           message.UserID,
		RecipientID:      message.RecipientID,
		Content:          message.Content,
		MessageType:      message.MessageType,
		IsPinned:         message.IsPinned,
		IsEdited:         message.IsEdited,
		EditedAt:         message.EditedAt,
		ReplyToMessageID: message.ReplyToMessageID,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
		User: dto.ChatUserResponse{
			ID:       uint64(message.User.ID),
			Username: message.User.Username,
			FullName: message.User.Name,
			Avatar:   s.generateAvatarURL(message.User.AvatarInfo),
		},
		Recipient: recipient,
	}

	if len(message.MessageMedias) > 0 {
		response.Medias = make([]dto.ChatMediaResponse, len(message.MessageMedias))
		for i, messageMedia := range message.MessageMedias {
			var fullStaticURL *string
			if messageMedia.Media.StaticURL != nil {
				diskName := "s3"
				if messageMedia.Media.DiskName != nil {
					diskName = *messageMedia.Media.DiskName
				}
				fullURL := utils.StaticURL(*messageMedia.Media.StaticURL, diskName)
				fullStaticURL = &fullURL
			}

			response.Medias[i] = dto.ChatMediaResponse{
				ID:            messageMedia.Media.ID,
				FileName:      messageMedia.Media.FileName,
				FilePath:      messageMedia.Media.FilePath,
				FullPath:      messageMedia.Media.FullPath,
				FileType:      messageMedia.Media.FileType,
				FileSize:      messageMedia.Media.FileSize,
				FileExtension: messageMedia.Media.FileExtension,
				DiskName:      messageMedia.Media.DiskName,
				StaticURL:     fullStaticURL,
				SortOrder:     messageMedia.SortOrder,
			}
		}
	}

	// Populate ReplyToMessage if this message is a reply
	if message.ReplyToMessage != nil {
		response.ReplyToMessage = &dto.ChatMessageResponse{
			ID:          message.ReplyToMessage.ID,
			CourseID:    message.ReplyToMessage.CourseID,
			UserID:      message.ReplyToMessage.UserID,
			Content:     message.ReplyToMessage.Content,
			MessageType: message.ReplyToMessage.MessageType,
			IsPinned:    message.ReplyToMessage.IsPinned,
			IsEdited:    message.ReplyToMessage.IsEdited,
			CreatedAt:   message.ReplyToMessage.CreatedAt,
			UpdatedAt:   message.ReplyToMessage.UpdatedAt,
			User: dto.ChatUserResponse{
				ID:       uint64(message.ReplyToMessage.User.ID),
				Username: message.ReplyToMessage.User.Username,
				FullName: message.ReplyToMessage.User.Name,
				Avatar:   s.generateAvatarURL(message.ReplyToMessage.User.AvatarInfo),
			},
		}
	}

	return response
}

func (s *chatMessageService) determineMessageTypeFromMedias(ctx context.Context, mediaIDs []int64) string {
	if len(mediaIDs) == 0 {
		return "text"
	}

	var fileTypes []string
	for _, mediaID := range mediaIDs {
		media, err := s.mediaRepo.FindById(mediaID)
		if err != nil {
			continue
		}

		if media.FileExtension == nil {
			fileTypes = append(fileTypes, "file")
			continue
		}

		ext := strings.ToLower(*media.FileExtension)

		// Images
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		isImage := false
		for _, imgExt := range imageExts {
			if ext == imgExt {
				fileTypes = append(fileTypes, "image")
				isImage = true
				break
			}
		}
		if isImage {
			continue
		}

		// Videos
		videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
		isVideo := false
		for _, vidExt := range videoExts {
			if ext == vidExt {
				fileTypes = append(fileTypes, "video")
				isVideo = true
				break
			}
		}
		if isVideo {
			continue
		}

		// Audio
		audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
		isAudio := false
		for _, audExt := range audioExts {
			if ext == audExt {
				fileTypes = append(fileTypes, "audio")
				isAudio = true
				break
			}
		}
		if isAudio {
			continue
		}

		// Default to file
		fileTypes = append(fileTypes, "file")
	}

	if len(fileTypes) > 0 {
		firstType := fileTypes[0]
		allSameType := true
		for _, t := range fileTypes {
			if t != firstType {
				allSameType = false
				break
			}
		}
		if allSameType {
			return firstType
		}
	}

	return "mixed"
}

func (s *chatMessageService) determineMessageTypeFromFiles(files []*multipart.FileHeader) string {
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))

		// Images
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		for _, imgExt := range imageExts {
			if ext == imgExt {
				return "image"
			}
		}

		// Videos
		videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
		for _, vidExt := range videoExts {
			if ext == vidExt {
				return "video"
			}
		}

		// Audio
		audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
		for _, audExt := range audioExts {
			if ext == audExt {
				return "audio"
			}
		}
	}

	return "file"
}

func (s *chatMessageService) generateAvatarURL(avatarInfo models.MediaInfo) string {
	if avatarInfo.Path == "" {
		return ""
	}

	avatarURL := utils.StaticURL(avatarInfo.Path, avatarInfo.Disk)

	return avatarURL
}

func (s *chatMessageService) SendToRecipient(c *gin.Context, courseId, senderId, recipientId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	ctx := c.Request.Context()
	s.userRepo.SetContext(c)
	recipient, err := s.userRepo.FindByID(int(recipientId))
	if err != nil || recipient == nil {
		return nil, fmt.Errorf("recipient not found: %w", err)
	}

	sender, err := s.userRepo.FindByID(int(senderId))
	if err != nil || sender == nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("message must have either content or media attachments")
	}
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	messageType := "text"
	if len(req.MediaIDs) > 0 {
		messageType = s.determineMessageTypeFromMedias(ctx, req.MediaIDs)
	}

	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      senderId,
		RecipientID: &recipientId,
		Content:     &req.Content,
		MessageType: messageType,
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.SendToRecipient(ctx, message, req.MediaIDs); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	fullMessage, err := s.repo.GetByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	response := s.mapToResponse(fullMessage)

	// Notify private message event in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewPrivateMessage: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewPrivateMessage(courseId, message.ID, senderId, recipientId, response); err != nil {
			config.Log.Warn("⚠️ Failed to publish new private message notification: %v", err)
		}
	}()

	// Notify recipient about new message for recent senders update
	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyUser new_private_message: %v", r)
			}
		}()

		// Prepare notification data for recent senders update
		notificationData := map[string]interface{}{
			"course_id":  courseId,
			"message_id": message.ID,
			"sender_id":  senderId,
			"sender": dto.ChatUserResponse{
				ID:       uint64(sender.ID),
				Username: sender.Username,
				FullName: sender.Name,
				Avatar:   s.generateAvatarURL(sender.AvatarInfo),
			},
			"last_message": response,
		}

		if err := s.pubsubService.NotifyUser(recipientId, "new_private_message", notificationData); err != nil {
			config.Log.Warn("⚠️ Failed to publish user notification for new private message: %v", err)
		}
	}()

	return response, nil
}

func (s *chatMessageService) GetMessagesWithRecipient(c *gin.Context, courseId, senderId, recipientId uint64, page, limit int) (*dto.ChatMessagesListResponse, error) {
	ctx := c.Request.Context()
	messages, total, err := s.repo.GetMessagesWithRecipient(ctx, courseId, senderId, recipientId, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	responseMessages := make([]dto.ChatMessageResponse, 0, len(messages))
	for i := range messages {
		responseMessages = append(responseMessages, *s.mapToResponse(&messages[i]))
	}

	totalPage := (int(total) + limit - 1) / limit
	hasMore := page < totalPage

	return &dto.ChatMessagesListResponse{
		Messages: responseMessages,
		Pagination: dto.PaginationResponse{
			Page:      page,
			Limit:     limit,
			Total:     int(total),
			TotalPage: totalPage,
			HasMore:   hasMore,
		},
	}, nil
}

func (s *chatMessageService) GetRecentSenders(c *gin.Context, courseId, currentUserId uint64, limit int) ([]dto.RecentSenderResponse, error) {
	ctx := c.Request.Context()
	s.userRepo.SetContext(c)
	messages, err := s.repo.GetRecentSenders(ctx, courseId, currentUserId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent senders: %w", err)
	}

	results := make([]dto.RecentSenderResponse, 0, len(messages))
	for i := range messages {
		msg := &messages[i]

		var sender *models.User
		if msg.UserID == currentUserId {
			if msg.Recipient != nil {
				sender = msg.Recipient
			} else if msg.RecipientID != nil {
				recipient, err := s.userRepo.FindByID(int(*msg.RecipientID))
				if err == nil {
					sender = recipient
				}
			}
		} else {
			user, err := s.userRepo.FindByID(int(msg.UserID))
			if err == nil {
				sender = user
			}
		}

		if sender == nil {
			continue
		}

		lastMessage := s.mapToResponse(msg)

		senderResponse := dto.ChatUserResponse{
			ID:       uint64(sender.ID),
			Username: sender.Username,
			FullName: sender.Name,
			Avatar:   s.generateAvatarURL(sender.AvatarInfo),
		}

		results = append(results, dto.RecentSenderResponse{
			Sender:      senderResponse,
			LastMessage: *lastMessage,
		})
	}

	return results, nil
}
